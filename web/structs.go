package web

type checkSuite struct {
	AfterSHA   string `json:"after"`
	BeforeSHA  string `json:"before"`
	Conclusion string `json:"conclusion"`
	HeadBranch string `json:"head_branch"`
	HeadCommit string `json:"head_commit"`
	HeadSHA    string `json:"head_sha"`
	Status     string `json:"status"`
}

type checkRun struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Conclusion string `json:"conclusion"`
	DetailsURL string `json:"details_url"`
	StartedAt  string `json:"started_at"`
	Status     string `json:"status"`
	URL        string `json:"url"`

	Output struct {
		AnnotationsURL string `json:"annotations_url"`
		Summary        string `json:"summary"`
		Text           string `json:"text"`
		Title          string `json:"title"`
	} `json:"output"`

	CheckSuite struct {
		checkSuite
		Repository repository `json:"repository"`
	} `json:"check_suite"`
	Deployment deployment `json:"deployment"`
}

type repository struct {
	Name           string `json:"name"`
	URL            string `json:"URL"`
	CloneURL       string `json:"clone_url"`
	StatusesURL    string `json:"statuses_url"`
	DeploymentsURL string `json:"deployments_url"`

	App struct {
		ExternalURL string `json:"external_url"`
	} `json:"app"`
}

type deployment struct{}

type checkSuiteRequest struct {
	CheckSuite checkSuite `json:"check_suite"`
	Repository repository `json:"repository"`
}

type checkRunRequest struct {
	CheckRun checkRun `json:"check_run"`
}
