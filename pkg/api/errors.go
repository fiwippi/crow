package api

import "errors"

var (
	ErrInvalidScheme = errors.New("invalid scheme")
	ErrEndpointType = errors.New("endpoint is not one for api, media, static content")
	ErrInvalidPageNum = errors.New("invalid page num, should be in the range 1-15 inclusive")
	ErrInvalidSpoilerNum = errors.New("invalid spoiler num, should be in the range 1-5 inclusive")
	ErrNotFound = errors.New("resource not found")
	ErrInvalidAssetFormat = errors.New("assets should be formatted as 'folder/name.ext' or 'folder/name.num.ext'")
)
