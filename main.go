// Copyright 2018 AMIS Technologies
// This file is part of the sol2proto
//
// The sol2proto is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The sol2proto is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the sol2proto. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/getamis/sirius/util"
	flag "github.com/spf13/pflag"

	"github.com/getamis/sol2proto/grpc"
)

var (
	GitVersion string = "dirty"
	abiFiles   []string
	pkgName    string
	output     string
)

func init() {
	flag.StringArrayVar(&abiFiles, "abi", []string{}, "ABI files generated by solc")
	flag.StringVar(&pkgName, "pkg", "pb", "go package name for the generated proto")
	flag.StringVarP(&output, "output", "o", "stdout", "Output destination")
}

func main() {
	flag.Parse()

	if len(abiFiles) == 0 {
		fmt.Printf("Please specify the abi files\n")
		os.Exit(-1)
	}

	if pkgName == "" {
		fmt.Printf("Please specify package name\n")
		os.Exit(-1)
	}

	var serviceProtos []grpc.ProtoFile
	var requiredMsgs []grpc.Message

	for _, f := range abiFiles {
		abiString, err := ioutil.ReadFile(f)
		if err != nil {
			fmt.Printf("Failed to read input ABI: %v\n", err)
			os.Exit(-1)
		}

		contractAbi, err := abi.JSON(bytes.NewReader(abiString))
		if err != nil {
			fmt.Printf("Failed to parse contract ABI: %v\n", err)
			os.Exit(-1)
		}

		srvName := util.ToCamelCase(strings.TrimSuffix(filepath.Base(f), filepath.Ext(filepath.Base(f))))
		proto, msgs := grpc.GenerateServiceProtoFile(srvName, pkgName, contractAbi, GitVersion)

		serviceProtos = append(serviceProtos, proto)
		requiredMsgs = append(requiredMsgs, msgs...)
	}

	// generate messages proto file
	msgProto := grpc.GenerateMessageProtoFile("Messages", pkgName, abiFiles, requiredMsgs, GitVersion)

	writeCloser, needClose := getDestinationWriter(msgProto.Name)
	if needClose {
		defer writeCloser.Close()
	}
	err := msgProto.Render(writeCloser)
	if err != nil {
		fmt.Printf("Failed to render %s, %v\n", msgProto.Name, err)
		os.Exit(-1)
	}

	// generate service proto files
	for _, srvProto := range serviceProtos {
		writeCloser, needClose := getDestinationWriter(srvProto.Name)
		if needClose {
			defer writeCloser.Close()
		}
		err := srvProto.Render(writeCloser)
		if err != nil {
			fmt.Printf("Failed to render %s, %v\n", srvProto.Name, err)
			os.Exit(-1)
		}
	}
}

func getDestinationWriter(filename string) (destination io.WriteCloser, needClose bool) {
	var err error
	filename = fmt.Sprintf("%s.proto", util.ToUnderScore(filename))

	switch output {
	case "stdout":
		destination = os.Stdout
		needClose = false
	case "stderr":
		destination = os.Stderr
		needClose = false
	case "file":
		fallthrough
	default:
		destination, err = os.Create(filename)
		if err != nil {
			fmt.Printf("Failed to create file %s, %v\n", filename, err)
			os.Exit(-1)
		}

		needClose = true
	}

	return destination, needClose
}
