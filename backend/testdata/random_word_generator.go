package testdata

import (
	"math/rand"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type WordList struct {
	file  string
	words []string
	rand  *rand.Rand
	index int
}

func NewWordList(file string, rand *rand.Rand) *WordList {
	return &WordList{
		file: file,
		rand: rand,
	}
}

func (w *WordList) Load() error {
	words, err := os.ReadFile(w.file)
	if err != nil {
		return err
	}

	w.words = strings.Fields(string(words))
	w.Shuffle()

	return nil
}

func (w *WordList) RandomWord() string {
	if w.words == nil {
		panic("word list not loaded")
	}

	if w.index >= len(w.words) {
		w.Shuffle()
	}

	word := w.words[w.index]
	w.index++

	return word
}

func (w *WordList) Shuffle() {
	w.rand.Shuffle(len(w.words), func(i, j int) {
		w.words[i], w.words[j] = w.words[j], w.words[i]
	})
	w.index = 0
}

type RandomWordGenerator struct {
	rand      *rand.Rand
	wordlists map[string]*WordList
}

func NewRandomWordGenerator(rand *rand.Rand) *RandomWordGenerator {
	return &RandomWordGenerator{
		rand:      rand,
		wordlists: make(map[string]*WordList),
	}
}

func (r *RandomWordGenerator) AddWordList(tag string, file string) {
	wordlist := NewWordList(file, r.rand)
	err := wordlist.Load()
	if err != nil {
		panic(err)
	}
	r.wordlists[tag] = wordlist
}

func (r *RandomWordGenerator) Word(listTag string) string {
	wordlist := r.wordlists[listTag]

	return wordlist.RandomWord()
}

func (r *RandomWordGenerator) CapitalizedWord(listTag string) string {
	word := r.Word(listTag)
	return cases.Title(language.English).String(word)
}

func (r *RandomWordGenerator) ResetUsed(listTag string) {
	wordlist := r.wordlists[listTag]
	wordlist.Shuffle()
}
