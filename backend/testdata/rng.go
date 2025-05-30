package testdata

import (
	"math/rand"
	"os"
	"path/filepath"
)

const (
	CharsetLowercase = "abcdefghijklmnopqrstuvwxyz"
	CharsetUppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetDigits    = "0123456789"
	CharsetSymbols   = "!@#$%^&*()_+-=[]{}|;:,.<>?~"
	CharsetLetters   = CharsetLowercase + CharsetUppercase
	CharsetAlphanum  = CharsetLetters + CharsetDigits
	CharsetAll       = CharsetAlphanum + CharsetSymbols
)

type Rng struct {
	*rand.Rand
	wordGenerator *RandomWordGenerator
}

func NewRng(seed int64) *Rng {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	rand := rand.New(rand.NewSource(seed))

	wordGenerator := NewRandomWordGenerator(rand)
	wordGenerator.AddWordList("noun", filepath.Join(wd, "testdata/wordlists/top_english_nouns_lower_10000.txt"))
	wordGenerator.AddWordList("adj", filepath.Join(wd, "testdata/wordlists/top_english_adjs_lower_10000.txt"))

	return &Rng{
		Rand:          rand,
		wordGenerator: wordGenerator,
	}
}

func (r *Rng) Noun() string {
	return r.wordGenerator.Word("noun")
}

func (r *Rng) Adjective() string {
	return r.wordGenerator.Word("adj")
}

func (r *Rng) CapitalizedNoun() string {
	return r.wordGenerator.CapitalizedWord("noun")
}

func (r *Rng) CapitalizedAdjective() string {
	return r.wordGenerator.CapitalizedWord("adj")
}

func (r *Rng) String(out []byte, charset string) {
	for i := range out {
		out[i] = charset[r.Intn(len(charset))]
	}
}
