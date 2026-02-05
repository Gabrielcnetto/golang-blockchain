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
}

func (cli *CommandLine) PrintUsage() {
	fmt.Println("Usage: ")
	fmt.Println("getBalance -address ADDRESS - Get Balance for address")
	fmt.Println("createBlockchain  --address -Mine genesis block")
	fmt.Println("Send --from FROM --To TO --Amount AMOUNT - send to account")
	fmt.Println("printChain - Print all blocks from chain")

}

func (cli *CommandLine) PrintChain() {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	iter := chain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("Prev hash:%x\n", block.PrevHash)
		fmt.Printf("Hash:%x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Printf("pow: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		if len(block.PrevHash) == 0 {
			break
		}
	}
}
func (cli *CommandLine) CreateBlockchain(address string) {
	chain := blockchain.InitBlockchain(address)
	chain.Database.Close()
	fmt.Println("Finished")
}

func (cli *CommandLine) GetBalance(address string) {
	chain := blockchain.ContinueBlockchain(address)
	defer chain.Database.Close()
	balance := 0
	UTXOs := chain.FindUnspentTransactionOutput(address)
	for _, item := range UTXOs {
		balance += item.Value
	}
	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) Send(from, to string, amount int) {
	chain := blockchain.ContinueBlockchain(from)
	defer chain.Database.Close()
	tx := blockchain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Sucess!")
}
func (cli *CommandLine) Run() {
	cli.ValidateArgs()
	//get balance init
	getBalanceCMD := flag.NewFlagSet("getBalance", flag.ExitOnError)
	getBalanceAddres := getBalanceCMD.String("address", "", "Adress wallet")

	//get balance end

	//createblockchain init
	createBlockchainCMD := flag.NewFlagSet("createBlockchain", flag.ExitOnError)
	createBlockchainAdress := createBlockchainCMD.String("address", "", "address of miner")
	//create blockchain end
	//send init
	SendCmd := flag.NewFlagSet("Send", flag.ExitOnError)
	sendFrom := SendCmd.String("from", "", "from")
	sendTo := SendCmd.String("to", "", "to")
	amount := SendCmd.Int("amount", 0, "amount")
	//send end
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)

	switch os.Args[1] {
	case "getBalance":
		err := getBalanceCMD.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "createBlockchain":
		err := createBlockchainCMD.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "printChain":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)

	case "Send":
		err := SendCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	default:
		cli.PrintUsage()
		runtime.Goexit()
	}

	if getBalanceCMD.Parsed() {
		if *getBalanceAddres == "" {
			getBalanceCMD.Usage()
			runtime.Goexit()
		}
		cli.GetBalance(*getBalanceAddres)
	}
	if createBlockchainCMD.Parsed() {
		if *createBlockchainAdress == "" {
			createBlockchainCMD.Usage()
			runtime.Goexit()
		}
		cli.CreateBlockchain(*createBlockchainAdress)
	}
	if printChainCmd.Parsed() {
		cli.PrintChain()
	}
	if SendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *amount == 0 {
			SendCmd.Usage()
			runtime.Goexit()
		}
		cli.Send(*sendFrom, *sendTo, *amount)
	}
}
func (cli *CommandLine) ValidateArgs() {
	if len(os.Args) < 2 {
		cli.PrintUsage()
		runtime.Goexit()
	}
}

func main() {
	defer os.Exit(0)
	cli := CommandLine{}
	cli.Run()
}
