package links

type State string

const (
	StatePending State = "pending"
	StateAlive   State = "alive"
	StateFlaky   State = "flaky"
	StateDead    State = "dead"
)

type Outcome string

const (
	OutcomeOK           Outcome = "ok"
	OutcomeRedirectOK   Outcome = "redirect_ok"
	OutcomeBlocked      Outcome = "blocked"
	OutcomeNotFound     Outcome = "not_found"
	OutcomeGone         Outcome = "gone"
	OutcomeDNSFailure   Outcome = "dns_failure"
	OutcomeServerError  Outcome = "server_error"
	OutcomeTimeout      Outcome = "timeout"
	OutcomeConnRefused  Outcome = "conn_refused"
	OutcomeTLSError     Outcome = "tls_error"
	OutcomeRedirectLoop Outcome = "redirect_loop"
	OutcomeOtherError   Outcome = "other_error"
)

func (o Outcome) IsSuccess() bool {
	return o == OutcomeOK || o == OutcomeRedirectOK
}

func (o Outcome) IsBlocked() bool {
	return o == OutcomeBlocked
}

func (o Outcome) IsHardDeath() bool {
	return o == OutcomeNotFound || o == OutcomeGone || o == OutcomeDNSFailure
}

func (o Outcome) IsValid() bool {
	switch o {
	case OutcomeOK,
		OutcomeRedirectOK,
		OutcomeBlocked,
		OutcomeNotFound,
		OutcomeGone,
		OutcomeDNSFailure,
		OutcomeServerError,
		OutcomeTimeout,
		OutcomeConnRefused,
		OutcomeTLSError,
		OutcomeRedirectLoop,
		OutcomeOtherError:
		return true
	default:
		return false
	}
}
