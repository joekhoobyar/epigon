package storage_test

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github/joekhoobyar/api-emulator/storage"
	"github/joekhoobyar/api-emulator/test"
)

var _ = Describe("Fixtures", func() {
	var f *storage.FixtureStorage
	var root, child1, child2 []byte
	var err error

	dir := test.FixtureDir()

	BeforeEach(func() {
		f = storage.NewFixtureStorage(dir)

		root, err = os.ReadFile(path.Join(dir, "root.json"))
		Expect(err).NotTo(HaveOccurred())
		child1, err = os.ReadFile(path.Join(dir, "root/child1.json"))
		Expect(err).NotTo(HaveOccurred())
		child2, err = os.ReadFile(path.Join(dir, "root/child2.json"))
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Reading fixtures", func() {

		Context("that are objects", func() {
			It("should load root.json", func() { Expect(f.Read("root")).To(Equal(root)) })
			It("should load root/child1.json", func() { Expect(f.Read("root/child1")).To(Equal(child1)) })
			It("should load root/child2.json", func() { Expect(f.Read("root/child2")).To(Equal(child2)) })
		})

		Context("that are lists", func() {
			It("should report an error", func() {
				_, err := f.Read("root/")
				Expect(err).To(MatchError(HaveSuffix(" location does not identiy an object")))
			})
		})

		Context("that do not exist", func() {
			It("should report an error", func() {
				_, err = f.Read("missing")
				Expect(err).To(MatchError(HaveSuffix(" no such file or directory")))
			})
		})
	})

	Describe("Reading fixture lists", func() {
		Context("that are objects", func() {
			It("should report an error", func() {
				_, err = f.ReadList("root/child1")
				Expect(err).To(MatchError(HaveSuffix(" location does not identiy a collection")))
				_, err = f.ReadList("root")
				Expect(err).To(MatchError(HaveSuffix(" location does not identiy a collection")))
			})
		})

		Context("that are lists", func() {
			It("should load as an array of immediate children", func() {
				Expect(f.ReadList("root/")).To(Equal([]byte("[{\"name\":\"baby\"},{\"name\":\"kid\"}]")))
			})
		})
	})

	Describe("Listing fixture keys", func() {
		Context("that are objects", func() {
			It("should report an error", func() {
				_, err = f.List("root/child1")
				Expect(err).To(MatchError(HaveSuffix(" location does not identiy a collection")))
				_, err = f.List("root")
				Expect(err).To(MatchError(HaveSuffix(" location does not identiy a collection")))
			})
		})

		Context("that are lists", func() {
			It("should list subkeys of immediate children", func() {
				Expect(f.List("root/")).To(Equal([]string{"root/child1", "root/child2"}))
			})
		})
	})
})
