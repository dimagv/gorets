package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/jpfielding/gorets/cmds/common"
	"github.com/jpfielding/gorets/rets"
)

func main() {
	configFile := flag.String("config-file", "", "Config file for RETS connection")
	metadataFile := flag.String("metadata-options", "", "Config file for metadata options")
	output := flag.String("output", "", "Directory for file output")

	config := common.Config{}
	config.SetFlags()

	metadataOpts := MetadataOptions{}
	metadataOpts.SetFlags()

	flag.Parse()

	if *configFile != "" {
		err := config.LoadFrom(*configFile)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("Connection Settings: %v\n", config)
	if *metadataFile != "" {
		err := metadataOpts.LoadFrom(*metadataFile)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("Search Options: %v\n", metadataOpts)

	// should we throw an err here too?
	session, err := config.Initialize()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	capability, err := rets.Login(session, ctx, rets.LoginRequest{URL: config.URL})
	if err != nil {
		panic(err)
	}
	defer rets.Logout(session, ctx, rets.LogoutRequest{URL: capability.Logout})

	reader, err := rets.MetadataStream(session, ctx, rets.MetadataRequest{
		URL:    capability.GetMetadata,
		Format: metadataOpts.Format,
		MType:  metadataOpts.MType,
		ID:     metadataOpts.ID,
	})
	defer reader.Close()
	if err != nil {
		panic(err)
	}
	out := os.Stdout
	if *output != "" {
		out, _ = os.Create(*output + "/metadata.xml")
		defer out.Close()
	}
	io.Copy(out, reader)
}

// MetadataOptions ...
type MetadataOptions struct {
	MType  string `json:"metadata-type"`
	Format string `json:"format"`
	ID     string `json:"id"`
}

// SetFlags ...
func (o *MetadataOptions) SetFlags() {
	flag.StringVar(&o.MType, "mtype", "METADATA-SYSTEM", "The type of metadata requested")
	flag.StringVar(&o.Format, "format", "COMPACT", "Metadata format")
	flag.StringVar(&o.ID, "id", "*", "Metadata identifier")
}

// LoadFrom ...
func (o *MetadataOptions) LoadFrom(filename string) error {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	blob, err := ioutil.ReadAll(file)
	err = json.Unmarshal(blob, o)
	if err != nil {
		return err
	}
	return nil
}
