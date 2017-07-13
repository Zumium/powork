PoWork
======

A simple proof of work library for Golang
-----------------------------------------

__This is from bitbucket.org/RyanMarcus/powork originally__

** PoWork v0.1 **

PoWork provides a solution-verification based [proof of work system](http://en.wikipedia.org/wiki/Proof-of-work_system) that uses a probabilistic iteration method, similar to the ever-popular [Hashcash](http://hashcash.org). The library seeks to provide an abstracted interface to users seeking to use proof of work systems.

Full documentation can be found [on GoDoc](http://godoc.org/github.com/Zumium/powork).


You can import it with
	
	import "github.com/Zumium/powork"
	

To create a proof-of-work:

	worker := powork.NewWorker()
	messageToProve := "I'll prove I did some work with this very message!"

	// proof will contain both the message and the proof of work
	proof, _ := worker.DoProofForString(messageToProve)


To verify that proof-of-work:

    ok, _ := worker.ValidatePoWork(proof)
	if ok {
		fmt.Printf("The proven message is: %v\n", proof.GetMessageString())
	}


To change the proof difficulty (default: 10):

	// the default is 10. Increases are exponential!
	worker.SetDifficulty(15)

	// do a harder proof
	proof, _ = worker.DoProofForString(messageToProve)
	
To use a different hash function (default: SHA512):

	// use the MD5 hash function (for example)
	// the default is SHA512
	// you'll need to import MD5 with: import "crypto/md5"
	worker.SetHasher(md5.New())

	// do a proof with the MD5 hash function
	proof, _ = worker.DoProofForString(messageToProve)

To change the default timeout (default: 5 seconds)

	// set a new timeout value of 10 seconds (in milliseconds)
	worker.SetTimeout(10000)

	// now the worker will try for 10 seconds instead of 5
	proof, _ = worker.DoProofForString(messageToProve)

You can also use PoWork asynchronously by having PoWork return a channel:

	worker := NewWorker()
	messageToProve := []byte("This time with channels!")

	// returns a channel that will eventually get the proof
	recvr := worker.PrepareProof(messageToProve)

	// prepare your message here
	fmt.Printf("Preparing message...\n")


	// wait on the result
	res := <- recvr

	if res.error != nil {
		t.Fatalf("Error: %v\n", res.error)
	}

	proof := res.PoWork
	// this variable will contain the same thing as the
	// proof variables from the other examples

You can also have PoWork send multiple proofs to the same channel:

	worker := powork.NewWorker()
	messageToProve := []byte("Will send a proof of this message to a channel")
	
	c := powork.GetChannel(20)

	worker.SendProofToChannel(messageToProve, c)

	// do message preparation business here

	res := <- c
	
	if res.error != nil {
		t.Fatalf("Error occurred during proof generation: %v\n", res.error)
	}

	ok, _ := worker.ValidatePoWork(res.PoWork)
	if !ok {
		t.Fatalf("Could not validate message\n")
	}
