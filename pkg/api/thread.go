package api

func (t *Thread) addThreadAttributes() {
	t.No = t.Posts[0].No
	t.Closed = t.Posts[0].Closed
	t.Subject = t.Posts[0].Subject
	t.Comment = t.Posts[0].Comment
	t.Replies = t.Posts[0].Replies
	t.LastModified = t.Posts[0].LastModified
	t.UniqueIPs = t.Posts[0].UniqueIPs
	t.BumpLimit = t.Posts[0].BumpLimit
	t.ImageLimit = t.Posts[0].ImageLimit
	t.Archived = t.Posts[0].Archived
	t.ArchivedOn = t.Posts[0].ArchivedOn
	t.Sticky = t.Posts[0].Sticky
}
