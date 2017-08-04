package finalize

import (
	"github.com/blang/semver"
)

type Versions interface {
	HasGem(string) (bool, error)
	GemVersion(string) (*semver.Version, error)
}

func (f *Finalizer) GenerateReleaseYaml() (map[string]map[string]string, error) {
	processTypes := map[string]string{
		"rake":    "bundle exec rake",
		"console": "bundle exec irb",
	}
	hasThin, err := f.Versions.HasGem("thin")
	if err != nil {
		return nil, err
	}
	railsMM, err := f.Versions.GemVersion("rails")
	if err != nil {
		return nil, err
	}
	if railsMM != nil {
		rails := semver.MustParse(railsMM.String())
		processTypes["worker"] = "bundle exec rake jobs:work"
		if rails.GTE(mustParse("4.0.0-beta")) {
			processTypes["console"] = "bin/rails console"
			processTypes["web"] = "bin/rails server -b 0.0.0.0 -p $PORT -e $RAILS_ENV"
		} else if rails.GTE(semver.MustParse("3.0.0")) {
			processTypes["console"] = "bundle exec rails console"
			processTypes["web"] = "bundle exec rails server -p $PORT"
			if hasThin {
				processTypes["web"] = "bundle exec thin start -R config.ru -e $RAILS_ENV -p $PORT"
			}
		} else {
			processTypes["console"] = "bundle exec script/console"
			processTypes["web"] = "bundle exec ruby script/server -p $PORT"
			if hasThin {
				processTypes["web"] = "bundle exec thin start -e $RAILS_ENV -p $PORT"
			}
		}
	} else {
		hasRack, err := f.Versions.HasGem("rack")
		if err != nil {
			return nil, err
		}
		if hasRack {
			processTypes["web"] = "bundle exec rackup config.ru -p $PORT"
			if hasThin {
				processTypes["web"] = "bundle exec thin start -R config.ru -e $RACK_ENV -p $PORT"
			}
		}
	}
	return map[string]map[string]string{
		"default_process_types": processTypes,
	}, nil
}

func mustParse(s string) semver.Version {
	semver, err := semver.ParseTolerant(s)
	if err != nil {
		panic(err)
	}
	return semver
}