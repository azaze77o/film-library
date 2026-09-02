package domain

type ImdbID string

type MoviePreview struct {
	ID     ImdbID
	Title  string
	Year   string
	Type   string
	Poster string
}

type Movie struct {
	ID       ImdbID
	Title    string
	Year     string
	Genre    string
	Director string
	Actors   string
	Plot     string
	Country  string
	Ratings  []Rating
}

type Rating struct {
	Source string
	Value  string
}

type SearchResult struct {
	Movies       []MoviePreview
	TotalResults int
}
