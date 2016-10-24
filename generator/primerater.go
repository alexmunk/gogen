package generator

import (
	config "github.com/coccyx/gogen/internal"
	log "github.com/coccyx/gogen/logger"
	"github.com/coccyx/gogen/rater"
)

// PrimeRater ensures for a given sample, all raters are set for the tokens
func PrimeRater(s *config.Sample) {
	for i := 0; i < len(s.Tokens); i++ {
		t := s.Tokens[i]
		if t.Type == "rated" {
			if t.RaterString != "" && t.Rater == nil {
				log.Infof("Setting rater to %s for token '%s'", t.RaterString, t.Name)
				s.Tokens[i].Rater = rater.GetRater(t.RaterString)
			}
		}
	}
}
