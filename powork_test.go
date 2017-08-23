package powork

import "testing"
import "time"
import "crypto/md5"
import "golang.org/x/net/context"

func TestDefaultPoWork(t *testing.T) {
	toR := NewWorker()
	s := []byte("A test message")
	pow, err := toR.DoProofFor(s)

	if err != nil {
		t.Fatalf("An error occurred while calculating a default proof of work: %v\n", err)
	}

	var ok bool
	ok, err = toR.ValidatePoWork(pow)
	if err != nil {
		t.Fatalf("Created proof did not validate: %v\n", err)
	}

	if !ok {
		t.Fatalf("Proof did not validate.\n")
	}

	t.Logf("Required iterations: %v\n", pow.requiredIterations)
}

func TestDefaultPoWorkMany(t *testing.T) {
	testStrings := []string{
		"It is certain",
		"It is decidedly so",
		"Without a doubt",
		"Yes definitely",
		"You may rely on it",
		"As I see it yes",
		"Most likely",
		"Outlook good",
		"Yes",
		"Signs point to yes",
		"Reply hazy try again",
		"Ask again later",
		"Better not tell you now",
		"Cannot predict now",
		"Concentrate and ask again",
		"Don't count on it",
		"My reply is no",
		"My sources say no",
		"Outlook not so good",
		"Very doubtful",
	}

	toR := NewWorker()
	for _, s := range testStrings {

		pow, err := toR.DoProofFor([]byte(s))

		if err != nil {
			t.Fatalf("An error occurred while calculating a default proof of work: %v\n", err)
		}

		var ok bool
		ok, err = toR.ValidatePoWork(pow)
		if err != nil {
			t.Fatalf("Created proof did not validate: %v\n", err)
		}

		if !ok {
			t.Fatalf("Proof did not validate.\n")
		}

		t.Logf("Required iterations: %v\n", pow.requiredIterations)
	}
}

func TestDifficultyIncrease(t *testing.T) {
	toR := NewWorker()
	s := []byte("A test message for difficulty")
	pow, err := toR.DoProofFor(s)

	if err != nil {
		t.Fatalf("An error occurred while calculating a default proof of work: %v\n", err)
	}

	var ok bool
	ok, err = toR.ValidatePoWork(pow)
	if err != nil {
		t.Fatalf("Created proof did not validate: %v\n", err)
	}

	if !ok {
		t.Fatalf("Proof did not validate.\n")
	}

	req := pow.requiredIterations
	toR.SetDifficulty(16)
	pow, err = toR.DoProofFor(s)
	if err != nil {
		t.Fatalf("An error occurred while calculating a default proof of work: %v\n", err)
	}

	ok, err = toR.ValidatePoWork(pow)
	if err != nil {
		t.Fatalf("Created proof did not validate: %v\n", err)
	}

	if !ok {
		t.Fatalf("Proof did not validate.\n")
	}

	if pow.requiredIterations <= req {
		t.Fatalf("Increased difficulty did not require increased hash iterations")
	}

	t.Logf("Required iterations: %v\n", pow.requiredIterations)
}

func TestDefaultTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	toR := NewWorker()
	toR.SetDifficulty(100)
	s := []byte("A test message for difficulty")

	c := make(chan int, 1)
	go func(c chan int) {
		_, err := toR.DoProofFor(s)
		if err == nil {
			t.Fatalf("Difficulty of 100 did not produce a timeout.")
		}
		c <- 1
	}(c)

	beforeTimeout := time.After(time.Duration(toR.maxWait/2) * time.Millisecond)
	afterTimeout := time.After(time.Duration(toR.maxWait*2) * time.Millisecond)

	select {
	case <-c:
		// got timeout value
		t.Fatalf("Timed out too soon")
	case <-afterTimeout:
		// strange...
		t.Fatalf("Very unexpected solution to race condition. Try running again.")
	case <-beforeTimeout:
		// valid!
	}

	select {
	case <-c:
		// got timeout value
		// that's correct
	case <-afterTimeout:
		// strange...
		t.Fatalf("Did not timeout")

	}
}

func TestCustomTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	toR := NewWorker()
	toR.SetDifficulty(100)
	s := []byte("A test message for difficulty")

	c := make(chan int, 1)
	go func(c chan int) {
		_, err := toR.DoProofFor(s)
		if err == nil {
			t.Fatalf("Difficulty of 100 did not produce a timeout.")
		}
		c <- 1
	}(c)

	toR.SetTimeout(1000)

	beforeTimeout := time.After(time.Duration(toR.maxWait/2) * time.Millisecond)
	afterTimeout := time.After(time.Duration(toR.maxWait*2) * time.Millisecond)

	select {
	case <-c:
		// got timeout value
		t.Fatalf("Timed out too soon")
	case <-afterTimeout:
		// strange...
		t.Fatalf("Very unexpected solution to race condition. Try running again.")
	case <-beforeTimeout:
		// valid!
	}

	select {
	case <-c:
		// got timeout value
		// that's correct
	case <-afterTimeout:
		// strange...
		t.Fatalf("Did not timeout")

	}
}

func TestCustomHash(t *testing.T) {
	toR := NewWorker()
	toR.SetHasher(md5.New())
	// toR.SetHashGetter(md5.New)

	s := "A test message for MD5"
	res, err := toR.DoProofFor([]byte(s))
	if err != nil {
		t.Fatalf("Failed to construct proof with MD5: %v\n", err)
	}

	ok, verr := toR.ValidatePoWork(res)
	if verr != nil {
		t.Fatalf("Error while validating MD5 proof: %v\n", verr)
	}

	if !ok {
		t.Fatalf("MD5 proof of work failed to validate.\n")
	}

}

func TestAsync(t *testing.T) {
	worker := NewWorker()

	messageToProve := []byte("This time with channels!")

	recvr := worker.PrepareProof(messageToProve)

	// prepare your message here
	t.Logf("Preparing message...")

	// wait on the result
	res := <-recvr

	if res.error != nil {
		t.Fatalf("Error: %v\n", res.error)
	}

	proof := res.PoWork

	ok, _ := worker.ValidatePoWork(proof)

	if ok {
		t.Logf("The proven message is: %v\n", proof.GetMessageString())
	} else {
		t.Fatalf("Proof did not validate!")
	}

}

func TestSendToChannel(t *testing.T) {
	worker := NewWorker()
	messageToProve := []byte("Will send a proof of this message to a channel")

	c := GetChannel(20)

	worker.SendProofToChannel(messageToProve, c)

	res := <-c

	if res.error != nil {
		t.Fatalf("Error occurred during proof generation: %v\n", res.error)
	}

	ok, _ := worker.ValidatePoWork(res.PoWork)
	if !ok {
		t.Fatalf("Could not validate message\n")
	}

}

func TestMessageGetter(t *testing.T) {
	worker := NewWorker()
	messageToProve := "I'll prove I did some work with this very message!"

	proof, _ := worker.DoProofForString(messageToProve)
	ok, _ := worker.ValidatePoWork(proof)

	if ok {
		t.Logf("The proven message is: %v\n", proof.GetMessageString())
	} else {
		t.Fatalf("Proof did not validate!")
	}
}

func TestCancelCalculationPerpareProof(t *testing.T) {
	worker := NewWorker()
	messageToProve := []byte("I'll prove I did some work with this very message!")

	worker.SetDifficulty(30)

	ctx, cancelFunc := context.WithCancel(context.Background())
	out := worker.PrepareProofWithContext(ctx, messageToProve)
	<-time.After(time.Second)
	cancelFunc()
	r := <-out
	if r.error != context.Canceled {
		t.Fatal("error type wrong")
	}
}

func TestCancelCalculationSendProof(t *testing.T) {
	worker := NewWorker()
	messageToProve := []byte("I'll prove I did some work with this very message!")

	worker.SetDifficulty(30)

	out := make(chan struct {
		*PoWork
		error
	}, 1)
	defer close(out)
	ctx, cancelFunc := context.WithCancel(context.Background())
	worker.SendProofToChannelWithContext(ctx, messageToProve, out)
	<-time.After(time.Second)
	cancelFunc()
	r := <-out
	if r.error != context.Canceled {
		t.Fatal("error type wrong")
	}
}

func BenchmarkDefaultPoWork(b *testing.B) {
	worker := NewWorker()
	messageToProve := "I'll prove I did some work with this very message! Benchmark"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		proof, _ := worker.DoProofForString(messageToProve)
		ok, _ := worker.ValidatePoWork(proof)
		if !ok {
			b.Fatalf("Proof did not validate")
		}
	}
}

func BenchmarkIncreasedDifficulty(b *testing.B) {
	worker := NewWorker()
	worker.SetDifficulty(20)
	messageToProve := "I'll prove I did some work with this very message! Benchmark"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		proof, _ := worker.DoProofForString(messageToProve)
		ok, _ := worker.ValidatePoWork(proof)
		if !ok {
			b.Fatalf("Proof did not validate")
		}
	}
}

func BenchmarkAsyncProof(b *testing.B) {
	worker := NewWorker()
	messageToProve := []byte("Will send a proof of this message to a channel")

	c := GetChannel(20)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		worker.SendProofToChannel(messageToProve, c)
	}

	for i := 0; i < b.N; i++ {
		res := <-c

		if res.error != nil {
			b.Fatalf("Error occurred during proof generation: %v\n", res.error)
		}

		ok, _ := worker.ValidatePoWork(res.PoWork)
		if !ok {
			b.Fatalf("Could not validate message\n")
		}
	}

}
