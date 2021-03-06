/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	pb "kbe.grpctest/helloworld_mTLS/helloworld"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {

	//---------- TLS Setting -----------//
	//LoadX509KeyPair reads and parses a public/private key pair from a pair of files. 
	//The files must contain PEM encoded data. 
	//The certificate file may contain intermediate certificates following the leaf certificate to form a certificate chain.
	//On successful return, Certificate.Leaf will be nil because the parsed form of the certificate is not retained.
	certificate, err := tls.LoadX509KeyPair(
		//certificate signed by intermediary for the client. It contains the public key.
		"../cert_key/4_client/certs/localhost.cert.pem",
		//client key (needed to sign the requests, and only the public key can open the data)
		//If you encrypt data using someone’s public key, only their corresponding private key can decrypt it. 
		//On the other hand, if data is encrypted with the private key anyone can use the public key 
		//to unlock the message.
		"../cert_key/4_client/private/localhost.key.pem",
	)

	certPool := x509.NewCertPool()
	// chain is composed by ca.cert.pem and intermediate.cert.pem
	bs, err := ioutil.ReadFile("../cert_key/2_intermediate/certs/ca-chain.cert.pem")
	if err != nil {
		log.Fatalf("failed to read ca cert: %s", err)
	}
	//
        //AppendCertsFromPEM attempts to parse a series of PEM encoded certificates. 
	//It appends any certificates found to s and reports whether any certificates were successfully parsed.
        //On many Linux systems, /etc/ssl/cert.pem will contain the system wide set of root CAs 
	//in a format suitable for this function.
	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		log.Fatal("failed to append certs")
	}

	transportCreds := credentials.NewTLS(&tls.Config{
		// ServerName is used to verify the hostname on the returned
		// certificates unless InsecureSkipVerify is given. It is also included
		// in the client's handshake to support virtual hosting unless it is
		// an IP address.
		ServerName:   "localhost",
		// Credentials to present to the server
		// Certificates contains one or more certificate chains to present to the
	        // other side of the connection. The first certificate compatible with the
		// peer's requirements is selected automatically.
    		// Server configurations must set one of Certificates, GetCertificate or
  		// GetConfigForClient. Clients doing client-authentication may set either
                // Certificates or GetClientCertificate.
                //
                // Note: if there are multiple Certificates, and they don't have the
                // optional field Leaf set, certificate selection will incur a significant
                // per-handshake performance cost.
		Certificates: []tls.Certificate{certificate},
		// Safe store, trusted certificate list
		// Server need to use one certificate presents in this lists.
		// RootCAs defines the set of root certificate authorities
   		// that clients use when verifying server certificates.
   		// If RootCAs is nil, TLS uses the host's root CA set.
		RootCAs:      certPool,
	})


	//---------- gRPC Client -----------//
	dialOption := grpc.WithTransportCredentials(transportCreds)
	//conn, err := grpc.Dial(address, grpc.WithInsecure())
	conn, err := grpc.Dial(address, dialOption)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)
}
