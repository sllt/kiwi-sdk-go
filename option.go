package kiwi_sdk

type Option struct {
	Page    int
	Size    int
	Filters string
	Sort    string

	hackResponseRef any //hack for collection list
}
