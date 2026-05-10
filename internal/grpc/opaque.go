package grpc

import pb "github.com/amfib87/go-musthave-shortener-tpl/proto"

// Opaque‑обёртка для URLData
type URLDataOpaque struct {
	data *pb.URLData
}

func NewURLDataOpaque(data *pb.URLData) *URLDataOpaque {
	return &URLDataOpaque{data: data}
}

// Геттеры
func (u *URLDataOpaque) GetShortURL() string {
	return u.data.ShortUrl
}

func (u *URLDataOpaque) GetOriginalURL() string {
	return u.data.OriginalUrl
}

// Билдер для URLData
type URLDataBuilder struct {
	data *pb.URLData
}

func NewURLDataBuilder() *URLDataBuilder {
	return &URLDataBuilder{
		data: &pb.URLData{},
	}
}

func (b *URLDataBuilder) SetShortURL(shortURL string) *URLDataBuilder {
	b.data.ShortUrl = shortURL
	return b
}

func (b *URLDataBuilder) SetOriginalURL(originalURL string) *URLDataBuilder {
	b.data.OriginalUrl = originalURL
	return b
}

func (b *URLDataBuilder) Build() *pb.URLData {
	return b.data
}

// Opaque‑обёртка для UserURLsResponse
type UserURLsResponseOpaque struct {
	response *pb.UserURLsResponse
}

func NewUserURLsResponseOpaque(response *pb.UserURLsResponse) *UserURLsResponseOpaque {
	return &UserURLsResponseOpaque{response: response}
}

// Геттер для списка URL
func (r *UserURLsResponseOpaque) GetURLs() []*URLDataOpaque {
	opaqueURLs := make([]*URLDataOpaque, len(r.response.Url))
	for i, data := range r.response.Url {
		opaqueURLs[i] = NewURLDataOpaque(data)
	}
	return opaqueURLs
}

// Билдер для UserURLsResponse
type UserURLsResponseBuilder struct {
	response *pb.UserURLsResponse
}

func NewUserURLsResponseBuilder() *UserURLsResponseBuilder {
	return &UserURLsResponseBuilder{
		response: &pb.UserURLsResponse{},
	}
}

func (b *UserURLsResponseBuilder) AddURL(url *pb.URLData) *UserURLsResponseBuilder {
	b.response.Url = append(b.response.Url, url)
	return b
}

func (b *UserURLsResponseBuilder) AddURLOpaque(urlOpaque *URLDataOpaque) *UserURLsResponseBuilder {
	b.response.Url = append(b.response.Url, urlOpaque.data)
	return b
}

func (b *UserURLsResponseBuilder) Build() *pb.UserURLsResponse {
	return b.response
}

// Билдер для URLShortenResponse
type URLShortenResponseBuilder struct {
	response *pb.URLShortenResponse
}

func NewURLShortenResponseBuilder() *URLShortenResponseBuilder {
	return &URLShortenResponseBuilder{
		response: &pb.URLShortenResponse{},
	}
}

func (b *URLShortenResponseBuilder) SetResult(result string) *URLShortenResponseBuilder {
	b.response.Result = result
	return b
}

func (b *URLShortenResponseBuilder) Build() *pb.URLShortenResponse {
	return b.response
}
