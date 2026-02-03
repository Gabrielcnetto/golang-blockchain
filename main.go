package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/Gabrielcnetto/golang-blockchain/blockchain"
)

type CommandLine struct {
	blockchain *blockchain.Blockchain
}

func (cli *CommandLine) PrintUsage() {
	fmt.Println("Usage: ")
	fmt.Println("add -block BLOCK_DATA - add a block to the chain ")
	fmt.Println("print - Print all blocks from chain")

}
func (cli *CommandLine) AddBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("block added with sucess!")
}

func (cli *CommandLine) PrintChain() {
	iter := cli.blockchain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("Prev hash:%x\n", block.PrevHash)
		fmt.Printf("Data:%x\n", block.Data)
		fmt.Printf("Hash:%x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Printf("pow: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) Run() {
	cli.ValidateArgs()
	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")
	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	default:
		cli.PrintUsage()
		runtime.Goexit()
	}
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.AddBlock(*addBlockData)
	}
	if printChainCmd.Parsed() {
		cli.PrintChain()
	}
}
func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		runtime.Goexit()
	}
}

func main() {
	chain := blockchain.InitBlockchain()
	defer chain.Database.Close()
	cli := CommandLine{blockchain: chain}
	cli.Run()
}
