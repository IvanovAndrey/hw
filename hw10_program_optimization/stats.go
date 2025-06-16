package hw10programoptimization

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

//go:generate easyjson -all

type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

type users [100_000]User

func getUsers(r io.Reader) (result users, err error) {
	scanner := bufio.NewScanner(r)
	i := 0

	for scanner.Scan() {
		line := scanner.Bytes()

		var user User
		if err = user.UnmarshalJSON(line); err != nil {
			return
		}
		result[i] = user
		i++
	}

	if err = scanner.Err(); err != nil {
		return
	}

	return
}

func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)

	domain = strings.ToLower(domain)

	for _, user := range u {
		parts := strings.SplitN(user.Email, "@", 2)
		if len(parts) != 2 {
			continue
		}
		emailDomain := strings.ToLower(parts[1])

		if strings.HasSuffix(emailDomain, "."+domain) {
			result[emailDomain]++
		}
	}
	return result, nil
}

func GetDomainStatOld(r io.Reader, domain string) (DomainStat, error) {
	u, err := getUsersOld(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomainsOld(u, domain)
}

func getUsersOld(r io.Reader) (result users, err error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		var user User
		if err = json.Unmarshal([]byte(line), &user); err != nil {
			return
		}
		result[i] = user
	}
	return
}

func countDomainsOld(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)

	for _, user := range u {
		matched, err := regexp.Match("\\."+domain, []byte(user.Email))
		if err != nil {
			return nil, err
		}

		if matched {
			num := result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]
			num++
			result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])] = num
		}
	}
	return result, nil
}
