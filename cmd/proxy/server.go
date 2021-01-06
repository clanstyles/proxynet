package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"resnetworking/pkg/http"
	"resnetworking/pkg/nic"
	"resnetworking/pkg/proxy"
	"resnetworking/pkg/proxy/authorizor"
	"resnetworking/pkg/proxy/network"
	"resnetworking/pkg/sock"
	"resnetworking/services/user"
	"resnetworking/services/user_group"
	"strings"
	"sync"

	"github.com/gocql/gocql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"github.com/scylladb/gocqlx/migrate"
)

var (
	UserService      user.Service
	UserGroupService user_group.Service
)

var (
	socksProxyPort string
	httpProxyPort  string
	databaseHosts  []string
)

func init() {
	var dbhosts string

	flag.StringVar(&socksProxyPort, "socks-proxy-port", "8383", "proxy port as a string")
	flag.StringVar(&httpProxyPort, "http-proxy-port", "8484", "proxy port as a string")
	flag.StringVar(&dbhosts, "database-hosts", "127.0.0.1", "database hosts")
	flag.Parse()

	databaseHosts = strings.Split(dbhosts, ",")
}

func main() {
	log.Println("[proxy] connecting the database")

	var (
		ctx = context.Background()
	)

	cluster := gocql.NewCluster(databaseHosts...)
	cluster.ProtoVersion = 4
	cluster.Keyspace = "resnetworking"
	cluster.Consistency = gocql.One

	sess, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer sess.Close()

	// Try to do CQL migrations
	if err := migrate.Migrate(ctx, sess, "./migrations"); err != nil {
		log.Fatal(errors.Wrap(err, "failed to migrate cql"))
	}

	UserService = user.NewDatastore(sess)
	// UserService = user.NewMemDatastore()
	// UserGroupService = user_group.NewMemoryDatastore()
	UserGroupService = user_group.NewDatastore(sess)

	log.Println("[proxy] getting all listenable addresses")
	addrs, err := nic.GetAddresses()
	if err != nil {
		log.Fatal(err)
	}

	auths := []proxy.Authenticator{
		authorizor.NewCredential(UserService),
		authorizor.NewUserIP(UserService),
	}

	nps := []proxy.NetworkPolicy{
		network.NewBlacklist(UserGroupService),
	}

	p := proxy.New(auths, nps)
	wg := sync.WaitGroup{}

	for _, addr := range addrs {
		wg.Add(2)
		socksAddress := fmt.Sprintf("%s:%s", addr, socksProxyPort)
		httpAddress := fmt.Sprintf("%s:%s", addr, httpProxyPort)

		go func(addr string) {
			log.Printf("[http proxy] listening on %s", addr)

			if err := Listen(sock.New(p), addr); err != nil {
				log.Fatal(err)
			}
		}(socksAddress)

		go func(addr string) {
			log.Printf("[sock proxy] listening on %s", addr)

			if err := Listen(http.New(p), addr); err != nil {
				log.Fatal(err)
			}
		}(httpAddress)
	}

	wg.Wait()
}

func Listen(server proxy.Server, address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return errors.Wrap(err, "failed to listen on address")
	}

	if err := server.Listen(l); err != nil {
		return errors.Wrap(err, "failed to create proxy listener")
	}

	return nil
}
