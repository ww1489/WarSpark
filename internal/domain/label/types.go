package label

type Paging struct {
	Cursors Cursors `json:"paging,omitempty"`
}

type Cursors struct {
	After  string `json:"after,omitempty"`
	Before string `json:"before,omitempty"`
}

type Label struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IconURLs any    `json:"iconUrls,omitempty"`
}

type LabelListResponse struct {
	Items  []Label `json:"items"`
	Paging Paging  `json:"paging"`
}
