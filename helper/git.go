package helper

import (
	"gopkg.in/src-d/go-git.v4"
	"io"
	"sort"
	"time"
)

type timedHash struct {
	When time.Time
	Hash string
}

type timeSlice []timedHash

func (p timeSlice) Len() int {
	return len(p)
}

// Define compare
func (p timeSlice) Less(i, j int) bool {
	return p[i].When.Before(p[j].When)
}

// Define swap over an array
func (p timeSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func LatestCommitHash() (string, error) {
	r, err := git.PlainOpen(".")
	if err != nil {
		return "", err
	}

	iter, e := r.CommitObjects()
	if e != nil {
		return "", e
	}
	defer iter.Close()
	var hashes timeSlice
	for {
		commit, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		hashes = append(hashes, timedHash{commit.Author.When, commit.Hash.String()})
	}

	sort.Sort(hashes)
	return hashes[len(hashes)-1].Hash, nil
}
