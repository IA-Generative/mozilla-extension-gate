package sourceProviderLib

type Release struct {
	ID        string
	Name      string
	Assets    *[]ReleaseAsset
	AssetsURL string
}
type ReleaseAsset struct {
	ID   string
	Name string
	URL  string
	Size int64
	ContentType string
}