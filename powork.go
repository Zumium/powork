//Package powork provides an easy-to-use proof of work library for golang
package powork

import (
	"encoding/binary"
	"errors"
	"hash"
	"time"

	"golang.org/x/crypto/sha3"
)

// Worker represents an object that calculates proofs of work and verifies them.
type Worker struct {
	difficulty int
	// getHash    func() hash.Hash
	hasher  hash.Hash
	maxWait int
}

// A PoWork represents a (potentially valid) proof of work for a given message
type PoWork struct {
	msg                []byte
	proof              uint64
	requiredIterations int
}

// GetChannel returns a channel, with the given buffer, that can be used with SendProofToChannel
func GetChannel(buffer int) chan struct {
	*PoWork
	error
} {
	return make(chan struct {
		*PoWork
		error
	}, buffer)
}

// GetMessage gets the message that the proof of work relates to
func (p *PoWork) GetMessage() []byte {
	return p.msg
}

// GetMessageString simply casts the result of GetMessage to a string
func (p *PoWork) GetMessageString() string {
	return string(p.msg)
}

// NewWorker creates a new Worker with sensible defaults: SHA3-512, 10 bit difficulty, and a 5 second timeout.
func NewWorker() *Worker {
	return NewWorkerWithHash(sha3.New512()) // SHA3-512 by default
}

// NewWorkerWithHash creates a new worker with given hash
func NewWorkerWithHash(h hash.Hash) *Worker {
	w := new(Worker)
	w.difficulty = 10
	w.hasher = h
	w.maxWait = 5000
	return w
}

// SetDifficulty sets the difficulty of the proof calculated. A higher value represents a more difficult proof. Increases exponentially.
func (p *Worker) SetDifficulty(difficulty int) error {
	p.difficulty = difficulty
	if difficulty <= 0 {
		return errors.New("Difficulty must be at least 1")
	}

	return nil
}

// SetTimeout sets the amount of time a Worker will spend computing a proof of work before giving up.
func (p *Worker) SetTimeout(milliseconds int) error {
	p.maxWait = milliseconds
	if p.maxWait < 0 {
		return errors.New("Timeout must be greater than or equal to 0")
	}

	return nil
}

// SetHasher sets the hash object that the Worker will use
func (p *Worker) SetHasher(h hash.Hash) {
	p.hasher = h
}

// PrepareProof starts working on creating a proof of work for the passed message and
// returns immediately.
func (p *Worker) PrepareProof(msg []byte) chan struct {
	*PoWork
	error
} {

	toR := make(chan struct {
		*PoWork
		error
	}, 1)

	go func() {
		r, e := p.DoProofFor(msg)
		toR <- struct {
			*PoWork
			error
		}{r, e}
		close(toR)
	}()

	return toR
}

// SendProofToChannel begins computing a proof of work for the given message, and sends it to
// the passed channel upon completion.
func (p *Worker) SendProofToChannel(msg []byte, c chan struct {
	*PoWork
	error
}) {
	go func() {
		r, e := p.DoProofFor(msg)
		c <- struct {
			*PoWork
			error
		}{r, e}
	}()
}

// DoProofForString calculates a proof of work for a given string
func (p *Worker) DoProofForString(msg string) (*PoWork, error) {
	return p.DoProofFor([]byte(msg))
}

// DoProofFor calculates a proof of work for a byte slice
func (p *Worker) DoProofFor(msg []byte) (*PoWork, error) {
	toR := new(PoWork)
	toR.msg = msg
	toR.proof = 0
	toR.requiredIterations = 0

	timeoutChannel := time.After(time.Duration(p.maxWait) * time.Millisecond)

	for {
		res, err := p.ValidatePoWork(toR)
		if err != nil {
			return nil, err
		}

		if res {
			break
		}
		toR.requiredIterations++
		toR.proof++

		select {
		case <-timeoutChannel:
			// timed out
			return nil, errors.New("Timed out while calculating proof of work")
		default:
			// continue with the next iteration of the loop
		}
	}

	return toR, nil
}

// ValidatePoWork checks the validity of a proof of work. If the proof is valid,
// true is returned. Otherwise, false. If true is returned, then the
// error returned must be nil.
func (p *Worker) ValidatePoWork(pow *PoWork) (bool, error) {
	// hash := p.getHash()

	p.hasher.Reset()
	_, err := p.hasher.Write(pow.msg)
	if err != nil {
		return false, err
	}

	err = binary.Write(p.hasher, binary.LittleEndian, pow.proof)
	if err != nil {
		return false, err
	}

	sum := p.hasher.Sum(nil)
	// validate that the first N bits of the sum are 0, where N = p.difficulty
	N := p.difficulty
	for _, x := range sum {
		for i := 0; i < 8; i++ {
			if (x<<1)>>1 == x {
				N--
				x = x << 1
			} else {
				return false, nil
			}

			if N == 0 {
				//fmt.Printf("Valid hash: %X\n", sum)
				return true, nil
			}
		}
	}

	return false, errors.New("Buffer overrun: not enough bits in hash")

}
