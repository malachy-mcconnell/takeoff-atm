package bank

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("USD", Label("USD"), func() {
	It("Should accept a valid string and return an integer representing the cents count", func() {
		var subject USD
		var err error

		subject, err = USDFromString("0")
		Expect(err).To(BeNil())
		Expect(subject).To(Equal(USD(0)))

		subject, err = USDFromString("5")
		Expect(err).To(BeNil())
		Expect(subject).To(Equal(USD(500)))

		subject, err = USDFromString("5.00")
		Expect(err).To(BeNil())
		Expect(subject).To(Equal(USD(500)))

		subject, err = USDFromString("1.00")
		Expect(err).To(BeNil())
		Expect(subject).To(Equal(USD(100)))

		subject, err = USDFromString("1.01")
		Expect(err).To(BeNil())
		Expect(subject).To(Equal(USD(101)))

		subject, err = USDFromString("1.001")
		Expect(err).To(BeNil())
		Expect(subject).To(Equal(USD(100)))

		subject, err = USDFromString("0.999")
		Expect(err).To(BeNil())
		Expect(subject).To(Equal(USD(100)))
	})

	It("Should return a string representing the dollar amount", func() {
		Expect(USD(0).ToString()).To(Equal("0.00"))
		Expect(USD(1045).ToString()).To(Equal("10.45"))
		Expect(USD(101045).ToString()).To(Equal("1010.45"))
	})

	It("Should return an error if the input string is non-numeric", func() {
		subject, err := USDFromString("Apple")
		Expect(subject).To(Equal(USD(0)))
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("invalid syntax"))
	})
})
