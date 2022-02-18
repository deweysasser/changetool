package repo

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/rs/zerolog/log"
	"sync"
)

type Repository struct {
	*git.Repository
	tags        map[string]plumbing.Hash
	reverseTags map[plumbing.Hash][]string
	filled      sync.Once
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

	return r.tags
}

func (r *Repository) ReverseTagMap() map[plumbing.Hash][]string {
	r.filled.Do(r.fillTags)

	return r.reverseTags
}

func (r *Repository) fillTags() {
	log.Debug().Msg("filling tags map")

	r.tags = make(map[string]plumbing.Hash)
	r.reverseTags = make(map[plumbing.Hash][]string)

	if tagRefs, err := r.Tags(); err != nil {
		log.Err(err).Msg("error reading tag references")
		return
	} else {

		if err := tagRefs.ForEach(func(t *plumbing.Reference) error {
			name := t.Name().Short()
			hash := t.Hash()

			r.tags[name] = hash
			r.reverseTags[hash] = append(r.reverseTags[hash], name)

			return nil
		}); err != nil {
			log.Err(err).Msg("error iterating tag references")
		}
	}

	if tagObjects, err := r.TagObjects(); err != nil {
		log.Err(err).Msg("error reading tag objects")
	} else {
		if err := tagObjects.ForEach(func(tag *object.Tag) error {
			name := tag.Name
			if commit, err := tag.Commit(); err != nil {
				log.Err(err).Str("tag_name", name).Msg("Error reading tag object")
			} else {
				hash := commit.Hash
				r.tags[name] = hash
				r.reverseTags[hash] = append(r.reverseTags[hash], name)
			}
			return nil
		}); err != nil {
			log.Err(err).Msg("error iterating tag references")
		}
	}
}
