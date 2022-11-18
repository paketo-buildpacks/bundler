package internal_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paketo-buildpacks/bundler/dependency/retrieval/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testReleaseFetcher(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect           = NewWithT(t).Expect
		fetcher          internal.ReleaseFetcher
		testReleaseIndex *httptest.Server
	)

	it.Before(func() {
		testReleaseIndex = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/not-here":
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintln(w, "")
				return
			case "/bad-start":
				fmt.Fprintln(w, ``)
				return
			case "/bad-end":
				fmt.Fprintln(w, badEndJSON)
				return
			case "/bad-json":
				fmt.Fprintln(w, badJSON)
				return
			case "/unexpected-json":
				fmt.Fprintln(w, unexpectedJSON)
				return
			case "/empty-index":
				fmt.Fprintln(w, `[]`)
				return
			default:
				fmt.Fprintln(w, indexJSON)
			}
		}))

		fetcher = internal.NewReleaseFetcher(testReleaseIndex.URL)
	})

	it.After(func() {
		testReleaseIndex.Close()
	})

	context("Get", func() {
		it("returns a slice of all releases that aren't prerelease from the index with needed subset of metadata", func() {
			releases, err := fetcher.Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(releases).To(Equal([]internal.Release{
				{
					Version:  "2.3.25",
					Licenses: []string{"MIT"},
					SHA256:   "fd81ec4635c4189b66fd0789537d5cb38b3810b70765f6e1e82dda15b97591ad",
				},
				{
					Version:  "2.3.24",
					Licenses: []string{"MIT"},
					SHA256:   "eaa2eb8c3892e870f979252b2196bd77eb551e1dbf3cdc4eb164ba01ec4438c4",
				},
			}))
		})
		context("error cases", func() {
			context("the release index URI is malformed", func() {
				it.Before(func() {
					fetcher = internal.NewReleaseFetcher(`/\|/invalid-garbage-uri`)
				})
				it("returns a descriptive error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError(ContainSubstring(`unsupported protocol scheme`)))
				})
			})
			context("the release index server returns an error status", func() {
				it.Before(func() {
					fetcher = internal.NewReleaseFetcher(fmt.Sprintf("%s/not-here", testReleaseIndex.URL))
				})
				it("returns a descriptive error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError("release index returned status 404"))
				})
			})
			context("the release index doesn't start with a proper JSON token", func() {
				it.Before(func() {
					fetcher = internal.NewReleaseFetcher(fmt.Sprintf("%s/bad-start", testReleaseIndex.URL))
				})
				it("returns a descriptive error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError(ContainSubstring("EOF")))
				})
			})
			context("the release index doesn't end with a proper JSON token", func() {
				it.Before(func() {
					fetcher = internal.NewReleaseFetcher(fmt.Sprintf("%s/bad-end", testReleaseIndex.URL))
				})
				it("returns a descriptive error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError(ContainSubstring("EOF")))
				})
			})
			context("the release index isn't a JSON array", func() {
				it.Before(func() {
					fetcher = internal.NewReleaseFetcher(fmt.Sprintf("%s/bad-json", testReleaseIndex.URL))
				})
				it("returns a descriptive error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError(ContainSubstring("not at beginning of value")))
				})
			})
			context("the release index returns JSON without required metadata", func() {
				it.Before(func() {
					fetcher = internal.NewReleaseFetcher(fmt.Sprintf("%s/unexpected-json", testReleaseIndex.URL))
				})
				it("returns a descriptive error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("release index element 1 missing version"))))
				})
			})
			context("the release index doesn't contain any releases", func() {
				it.Before(func() {
					fetcher = internal.NewReleaseFetcher(fmt.Sprintf("%s/empty-index", testReleaseIndex.URL))
				})
				it("returns a descriptive error", func() {
					_, err := fetcher.Get()
					Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("no valid releases found"))))
				})
			})
		})
	})
}

const indexJSON = `
[{
  "authors": "André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
  "built_at": "2022-11-02T00:00:00.000Z",
  "created_at": "2022-11-02T15:49:16.992Z",
  "description": "Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
  "downloads_count": 33718,
  "metadata": {
    "homepage_uri": "https://bundler.io/",
    "changelog_uri": "https://github.com/rubygems/rubygems/blob/master/bundler/CHANGELOG.md",
    "bug_tracker_uri": "https://github.com/rubygems/rubygems/issues?q=is%3Aopen+is%3Aissue+label%3ABundler",
    "source_code_uri": "https://github.com/rubygems/rubygems/tree/master/bundler"
  },
  "number": "2.3.25",
  "summary": "The best way to manage your application's dependencies",
  "platform": "ruby",
  "rubygems_version": ">= 2.5.2",
  "ruby_version": ">= 2.3.0",
  "prerelease": false,
  "licenses": [
    "MIT"
  ],
  "requirements": [],
  "sha": "fd81ec4635c4189b66fd0789537d5cb38b3810b70765f6e1e82dda15b97591ad"
},
{
  "authors": "André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
  "built_at": "2022-10-17T00:00:00.000Z",
  "created_at": "2022-10-17T12:48:57.839Z",
  "description": "Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
  "downloads_count": 4157767,
  "metadata": {
    "homepage_uri": "https://bundler.io/",
    "changelog_uri": "https://github.com/rubygems/rubygems/blob/master/bundler/CHANGELOG.md",
    "bug_tracker_uri": "https://github.com/rubygems/rubygems/issues?q=is%3Aopen+is%3Aissue+label%3ABundler",
    "source_code_uri": "https://github.com/rubygems/rubygems/tree/master/bundler"
  },
  "number": "2.3.24.rc5",
  "summary": "The best way to manage your application's dependencies",
  "platform": "ruby",
  "rubygems_version": ">= 2.5.2",
  "ruby_version": ">= 2.3.0",
  "prerelease": true,
  "licenses": [
    "MIT"
  ],
  "requirements": [],
  "sha": "eaa2eb8c3892e870f979252b2196bd77eb551e1dbf3cdc4eb164ba01ec4438c4"
},
{
  "authors": "André Arko, Samuel Giddins, Colby Swandale, Hiroshi Shibata, David Rodríguez, Grey Baker, Stephanie Morillo, Chris Morris, James Wen, Tim Moore, André Medeiros, Jessica Lynn Suttles, Terence Lee, Carl Lerche, Yehuda Katz",
  "built_at": "2022-10-17T00:00:00.000Z",
  "created_at": "2022-10-17T12:48:57.839Z",
  "description": "Bundler manages an application's dependencies through its entire life, across many machines, systematically and repeatably",
  "downloads_count": 4157767,
  "metadata": {
    "homepage_uri": "https://bundler.io/",
    "changelog_uri": "https://github.com/rubygems/rubygems/blob/master/bundler/CHANGELOG.md",
    "bug_tracker_uri": "https://github.com/rubygems/rubygems/issues?q=is%3Aopen+is%3Aissue+label%3ABundler",
    "source_code_uri": "https://github.com/rubygems/rubygems/tree/master/bundler"
  },
  "number": "2.3.24",
  "summary": "The best way to manage your application's dependencies",
  "platform": "ruby",
  "rubygems_version": ">= 2.5.2",
  "ruby_version": ">= 2.3.0",
  "prerelease": false,
  "licenses": [
    "MIT"
  ],
  "requirements": [],
  "sha": "eaa2eb8c3892e870f979252b2196bd77eb551e1dbf3cdc4eb164ba01ec4438c4"
}]`

const badEndJSON = `[{
	"number" : "1.2.3",
	"sha" : "abcde"
}`

const badJSON = `{
	"some_key" : "some value",
	"other_key" : 3
}`

const unexpectedJSON = `[{
	"number" : "1.2.3",
	"sha" : "abcde"
},
{
	"some-key" : "some value",
	"other-key" : 3
}]`
