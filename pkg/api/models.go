package api

import "encoding/json"

// general structs
type Post struct {
	// Custom fields implemented by crow
	Board   string `json:"board"`    // The directory the board is located in.
	HasFile bool   `json:"has_file"` // Whether the post has a file attached
	// Fields from the API
	No              int         `json:"no"`             // The numeric post ID
	RepliesTo       int         `json:"resto"`          // For replies: this is the ID of the thread being replied to. For OP: this value is zero
	Sticky          Bool        `json:"sticky"`         // If the thread is being pinned to the top of the page
	Closed          Bool        `json:"closed"`         // If the thread is closed to replies
	Now             string      `json:"now"`            // MM/DD/YY(Day)HH:MM (:SS on some boards), EST/EDT timezone
	Time            Timestamp   `json:"time"`           // UNIX timestamp the post was created
	Name            string      `json:"name"`           // Name user posted with. Defaults to Anonymous
	Trip            string      `json:"trip"`           // The user's tripcode, in format: !tripcode or !!securetripcode
	ID              string      `json:"id"`             // The poster's ID
	CapCode         string      `json:"cap_code"`       // The capcode identifier for a post
	Country         string      `json:"country"`        // Poster's ISO 3166-1 alpha-2 country code
	CountryName     string      `json:"country_name"`   // Poster's country name
	Subject         string      `json:"sub"`            // OP Subject text
	Comment         string      `json:"com"`            // Comment (HTML escaped)
	ImageID         json.Number `json:"tim"`            // Unix timestamp + microtime that an image was uploaded
	Filename        string      `json:"filename"`       // Filename as it appeared on the poster's device
	Ext             string      `json:"ext"`            // Filetype
	Filesize        int         `json:"fsize"`          // Size of uploaded file in bytes
	MD5             string      `json:"md5"`           // 24 character, packed base64 MD5 hash of file
	ImageWidth      int         `json:"w"`              // Image width dimension
	ImageHeight     int         `json:"h"`              // Image height dimension
	ThumbnailWidth  int         `json:"tn_w"`           // Thumbnail image width dimension
	ThumbnailHeight int         `json:"tn_h"`           // Thumbnail image height dimension
	FileDeleted     Bool        `json:"filedeleted"`    // If the file was deleted from the post
	ImageSpoiler    Bool        `json:"spoiler"`        // If the image was spoilered or not
	CustomSpoiler   int         `json:"custom_spoiler"` // The custom spoiler ID for a spoilered image
	OmittedPosts    int         `json:"omitted_posts"`  // Number of replies minus the number of previewed replies
	OmittedImages   int         `json:"omitted_images"` // Number of image replies minus the number of previewed image replies
	Replies         int         `json:"replies"`        // Total number of replies to a thread
	Images          int         `json:"images"`         // Total number of image replies to a thread
	BumpLimit       Bool        `json:"bump_limit"`     // If a thread has reached bumplimit, it will no longer bump
	ImageLimit      Bool        `json:"image_limit"`    // If an image has reached image limit, no more image replies can be made
	LastModified    Timestamp   `json:"last_modified"`  // The UNIX timestamp marking the last time the thread was modified (post added/modified/deleted, thread closed/sticky settings modified)
	Tag             string      `json:"tag"`            // The category of .swf upload
	SemanticURL     string      `json:"semantic_url"`   // SEO URL slug for thread
	Since4Pass      int         `json:"since4pass"`     // Year 4chan pass bought
	UniqueIPs       int         `json:"unique_ips"`     // Number of unique posters in a thread
	MImg            Bool        `json:"m_img"`          // Mobile optimized image exists for post
	LastReplies     []Post      `json:"last_replies"`   // JSON representation of the most recent replies to a thread
	Archived        Bool        `json:"archived"`       // Thread has reached the board's archive
	ArchivedOn      Timestamp   `json:"archived_on"`    // UNIX timestamp the post was archived
}

// boards.json
type Boards struct {
	Boards     []Board           `json:"boards"`
	TrollFlags map[string]string `json:"troll_flags"`
}

type Board struct {
	Board           string   `json:"board"`             // The directory the board is located in.
	Title           string   `json:"title"`             // The readable title at the top of the board.
	WorkSafe        Bool     `json:"ws_board"`          // Is the board worksafe
	PerPage         int      `json:"per_page"`          // How many threads are on a single index page
	Pages           int      `json:"pages"`             // How many index pages does the board have
	MaxFilesize     int      `json:"max_filesize"`      // Maximum file size allowed for non .webm attachments (in KB)
	MaxWebmFilesize int      `json:"max_webm_filesize"` // Maximum file size allowed for .webm attachments (in KB)
	MaxCommentChars int      `json:"max_comment_chars"` // Maximum number of characters allowed in a post comment
	MaxWebmDuration int      `json:"max_webm_duration"` // Maximum duration of a .webm attachment (in seconds)
	BumpLimit       int      `json:"bump_limit"`        // Maximum number of replies allowed to a thread before the thread stops bumping
	ImageLimit      int      `json:"image_limit"`       // Maximum number of image replies per thread before image replies are discarded
	Cooldowns       Cooldown `json:"cooldowns"`
	MetaDescription string   `json:"meta_description"` // SEO meta description content for a board
	Spoilers        Bool     `json:"spoilers"`         // Are spoilers enabled
	CustomSpoilers  int      `json:"custom_spoilers"`  // How many custom spoilers does the board have
	IsArchived      Bool     `json:"is_archived"`      // Are archives enabled for the board
	TrollFlags      Bool     `json:"troll_flags"`      // Are troll flags enabled on the board
	CountryFlags    Bool     `json:"country_flags"`    // Are flags showing the poster's country enabled on the board
	UserIDs         Bool     `json:"user_ids"`         // Are poster ID tags enabled on the board
	Oekaki          Bool     `json:"oekaki"`           // Can users submit drawings via browser the Oekaki app
	SjisTags        Bool     `json:"sjis_tags"`        // Can users submit sjis drawings using the [sjis] tags
	CodeTags        Bool     `json:"code_tags"`        // Board supports code syntax highlighting using the [code] tags
	MathTags        Bool     `json:"math_tags"`        // Board supports [math] TeX and [eqn] tags
	TextOnly        Bool     `json:"text_only"`        // Is image posting disabled for the board
	ForcedAnon      Bool     `json:"forced_anon"`      // Is the name field disabled on the board
	WebmAudio       Bool     `json:"webm_audio"`       // Are webms with audio allowed?
	RequireSubject  Bool     `json:"require_subject"`  // Do OPs require a subject
	MinImageWidth   int      `json:"min_image_width"`  // What is the minimum image width (in pixels)
	MinImageHeight  int      `json:"min_image_height"` // What is the minimum image height (in pixels)
}

type Cooldown struct {
	Threads int `json:"threads"`
	Replies int `json:"replies"`
	Images  int `json:"images"`
}

// threads.json
type ThreadList struct {
	Board string `json:"board"`
	Pages []ThreadListPage `json:"pages"`
}

type ThreadListPage struct {
	Page    int                `json:"page"`    // The page number that the following Threads array is on
	Threads []ThreadListThread `json:"threads"` // The array of ThreadListThread objects
}

type ThreadListThread struct {
	No           int       `json:"no"`            // The OP ID of a thread
	LastModified Timestamp `json:"last_modified"` // The UNIX timestamp marking the last time the thread was modified (post added/modified/deleted, thread closed/sticky settings modified)
	Replies      int       `json:"replies"`       // A numeric count of the number of replies in the thread
}

// catalog.json
type Catalog struct {
	Board string        `json:"board"`
	Pages []CatalogPage `json:"pages"`
}

type CatalogPage struct {
	Page    int            `json:"page"` // The page number that the following Threads array is on
	Threads CatalogThreads `json:"threads"`
}

type CatalogThreads []*Post

// archive.json
type Archive struct {
	Board string `json:"board"`
	PostIDs []int  `json:"post_ids"`// Array of integers. These are the OP numbers of archived threads
}

// [board]/[1-15].json, [board]/thread/[op ID].json
type Page struct {
	No      int      `json:"no"`
	Board   string   `json:"board"`
	Threads []Thread `json:"threads"`
}

type Thread struct {
	Board        string    `json:"board"`         // The board directory
	No           int       `json:"no"`            // The ID of the first post in the thread
	Closed       Bool      `json:"closed"`        // If the thread is closed to replies
	Subject      string    `json:"sub"`           // OP Subject text
	Comment      string    `json:"com"`           // Comment (HTML escaped)
	Replies      int       `json:"replies"`       // Total number of replies to a thread
	LastModified Timestamp `json:"last_modified"` // The UNIX timestamp marking the last time the thread was modified (post added/modified/deleted, thread closed/sticky settings modified)
	UniqueIPs    int       `json:"unique_ips"`    // Number of unique posters in a thread
	BumpLimit    Bool      `json:"bump_limit"`    // If a thread has reached bumplimit, it will no longer bump
	ImageLimit   Bool      `json:"image_limit"`   // If an image has reached image limit, no more image replies can be made
	Archived     Bool      `json:"archived"`      // Thread has reached the board's archive
	ArchivedOn   Timestamp `json:"archived_on"`   // UNIX timestamp the post was archived
	Sticky       Bool      `json:"sticky"`        // If the thread is being pinned to the top of the page
	Posts        []*Post   `json:"posts"`         // Posts in the thread
}
