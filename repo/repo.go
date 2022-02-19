package repo

import (
	"github.com/deweysasser/changetool/perf"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
	"sync"
)

type Repository struct {
	*git.Repository
	tagToCommitHash  map[string]plumbing.Hash
	commitHashToTags map[plumbing.Hash][]string
	filled           sync.Once
}

func New(path string) (*Repository, error) {
	if r, err := git.PlainOpen(path); err != nil {
		return nil, err
	} else {
		return &Repository{Repository: r}, nil
	}
}

func FromRepository(r *git.Repository, err error) (*Repository, error) {
	return &Repository{Repository: r}, err
}

func (r *Repository) TagMap() map[string]plumbing.Hash {
	r.filled.Do(r.fillTags)

	return r.tagToCommitHash
}

func (r *Repository) ReverseTagMap() map[plumbing.Hash][]string {
	r.filled.Do(r.fillTags)

	return r.commitHashToTags
}

func (r *Repository) fillTags() {
	defer perf.Timer("filling tagToCommitHash map").Stop()

	r.tagToCommitHash = make(map[string]plumbing.Hash)
	r.commitHashToTags = make(map[plumbing.Hash][]string)

	if tagRefs, err := r.Tags(); err != nil {
		log.Err(err).Msg("error reading tag references")
		return
	} else {

		if err := tagRefs.ForEach(func(t *plumbing.Reference) error {
			name := t.Name().Short()
			hash := t.Hash()

			// If it's a tag object, resolve it to the commit it's pointing to
			if tag, err := r.TagObject(hash); err == nil {
				hash = tag.Target
			}

			log.Debug().
				Str("tag", name).
				Str("hash", hash.String()[:6]).
				Msg("Examining tag")

			r.tagToCommitHash[name] = hash
			r.commitHashToTags[hash] = append(r.commitHashToTags[hash], name)

			return nil
		}); err != nil {
			log.Err(err).Msg("error iterating tag references")
		}
	}

}
