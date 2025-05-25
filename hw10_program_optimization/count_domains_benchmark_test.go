package hw10programoptimization

import (
	"testing"
)

func generateUsers(n int) users {
	var us users
	for i := 0; i < n; i++ {
		us[i] = User{Email: getTestEmail(i)}
	}
	return us
}

func getTestEmail(i int) string {
	if i%2 == 0 {
		return "user" + string(rune(i)) + "@example.com"
	}
	return "user" + string(rune(i)) + "@test.com"
}

func BenchmarkCountDomainsOriginal(b *testing.B) {
	us := generateUsers(10000)
	for i := 0; i < b.N; i++ {
		_, err := countDomainsOld(us, "com")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCountDomainsOptimized(b *testing.B) {
	us := generateUsers(10000)
	for i := 0; i < b.N; i++ {
		_, err := countDomains(us, "com")
		if err != nil {
			b.Fatal(err)
		}
	}
}
