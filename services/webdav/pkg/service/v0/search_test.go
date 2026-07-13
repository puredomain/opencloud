package svc

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/timestamppb"

	searchmsg "github.com/opencloud-eu/opencloud/protogen/gen/opencloud/messages/search/v0"
	"github.com/opencloud-eu/opencloud/services/webdav/pkg/propfind"
)

func TestSearch(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Search Suite")
}

var _ = Describe("SpacesSearchRegex", func() {
	DescribeTable("path matching",
		func(path, expectedSpace string, shouldMatch bool) {
			matches := spacesSearchRegex.FindStringSubmatch(path)
			if shouldMatch {
				Expect(matches).ToNot(BeNil(), "Expected path %q to match", path)
				Expect(matches[1]).To(Equal(expectedSpace), "Expected space to be %q", expectedSpace)
			} else {
				Expect(matches).To(BeNil(), "Expected path %q not to match", path)
			}
		},
		Entry("standard dav spaces path", "/dav/spaces/12345", "12345", true),
		Entry("remote.php dav spaces path", "/remote.php/dav/spaces/12345", "12345", true),
		Entry("standard dav spaces path with subpaths", "/dav/spaces/12345/some/folder", "12345", true),
		Entry("remote.php dav spaces path with subpaths", "/remote.php/dav/spaces/12345/some/folder", "12345", true),
		Entry("standard dav spaces path without space", "/dav/spaces/", "", false),
		Entry("remote.php dav spaces path without space", "/remote.php/dav/spaces/", "", false),
		Entry("prefix match only", "/dav/spaces", "", false),
		Entry("unrelated path", "/dav/files/123", "", false),
	)
})

// timocloud patch #3: photo/image/location facet props on search REPORT
// responses (appendPhotoProps).
var _ = Describe("AppendPhotoProps", func() {
	ptrS := func(s string) *string { return &s }
	ptrF := func(f float32) *float32 { return &f }
	ptrI := func(i int32) *int32 { return &i }
	ptrD := func(d float64) *float64 { return &d }

	propValue := func(ps *propfind.PropstatXML, local string) (string, bool) {
		for _, p := range ps.Prop {
			if p.XMLName.Local == local {
				return string(p.InnerXML), true
			}
		}
		return "", false
	}

	It("emits nothing when no facets are set", func() {
		ps := &propfind.PropstatXML{}
		appendPhotoProps(ps, &searchmsg.Entity{})
		Expect(ps.Prop).To(BeEmpty())
	})

	It("emits only the fields that are present", func() {
		ps := &propfind.PropstatXML{}
		appendPhotoProps(ps, &searchmsg.Entity{
			Photo: &searchmsg.Photo{CameraMake: ptrS("Apple")},
		})
		Expect(ps.Prop).To(HaveLen(1))
		v, ok := propValue(ps, "oc:photo-camera-make")
		Expect(ok).To(BeTrue())
		Expect(v).To(Equal("Apple"))
		_, ok = propValue(ps, "oc:photo-taken-date-time")
		Expect(ok).To(BeFalse())
	})

	It("emits the full facet set with correct formatting", func() {
		taken := time.Date(2026, 7, 4, 12, 30, 0, 0, time.UTC)
		ps := &propfind.PropstatXML{}
		appendPhotoProps(ps, &searchmsg.Entity{
			Photo: &searchmsg.Photo{
				CameraMake:    ptrS("Apple"),
				CameraModel:   ptrS("iPhone 15 Pro"),
				FNumber:       ptrF(1.8),
				FocalLength:   ptrF(6.9),
				Iso:           ptrI(125),
				Orientation:   ptrI(6),
				TakenDateTime: timestamppb.New(taken),
			},
			Image:    &searchmsg.Image{Width: ptrI(4032), Height: ptrI(3024)},
			Location: &searchmsg.GeoCoordinates{Latitude: ptrD(37.7749), Longitude: ptrD(-122.4194), Altitude: ptrD(16.5)},
		})
		expect := map[string]string{
			"oc:photo-taken-date-time": "2026-07-04T12:30:00Z",
			"oc:photo-camera-make":     "Apple",
			"oc:photo-camera-model":    "iPhone 15 Pro",
			"oc:photo-f-number":        "1.8",
			"oc:photo-focal-length":    "6.9",
			"oc:photo-iso":             "125",
			"oc:photo-orientation":     "6",
			"oc:image-width":           "4032",
			"oc:image-height":          "3024",
			"oc:location-latitude":     "37.7749",
			"oc:location-longitude":    "-122.4194",
			"oc:location-altitude":     "16.5",
		}
		Expect(ps.Prop).To(HaveLen(len(expect)))
		for name, want := range expect {
			v, ok := propValue(ps, name)
			Expect(ok).To(BeTrue(), "missing prop %s", name)
			Expect(v).To(Equal(want), "prop %s", name)
		}
	})
})
