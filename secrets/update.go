package secrets

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

const SECRET_DEFINITION_URLS = "https://raw.githubusercontent.com/gitleaks/gitleaks/master/config/gitleaks.toml"

type (
	GitLeaksSecretRule struct {
		Id          string `toml:"id"`
		Description string `toml:"description"`
		Regex       string `toml:"regex"`
	}

	GitLeaksSecretRules struct {
		Rules []GitLeaksSecretRule `toml:"rules"`
	}
)

func GetSecretsFromGitLeaks() (*GitLeaksSecretRules, error) {
	resp, err := http.Get(SECRET_DEFINITION_URLS)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	secretsData := &GitLeaksSecretRules{}
	if _, err := toml.Decode(string(body), secretsData); err != nil {
		return nil, err
	}
	return secretsData, nil
}

func UpdateSecretsFromGitLeaks() ([]byte, error) {
	gitLeaksSecrets, err := GetSecretsFromGitLeaks()
	if err != nil {
		return nil, err
	}

	currentRules := make(map[string]SecretGeneratorRule)
	for _, gitLeaksSecret := range gitLeaksSecrets.Rules {
		currentRules[gitLeaksSecret.Id] = SecretGeneratorRule{
			Name:        gitLeaksSecret.Id,
			NameRegex:   formatSecretsNameRegex(gitLeaksSecret.Id),
			SecretRegex: formatSecretsRegex(gitLeaksSecret.Regex),
			Weight:      1,
		}
	}

	secretRulesYaml, err := yaml.Marshal(currentRules)
	if err != nil {
		return nil, err
	}

	return secretRulesYaml, nil
}

func formatSecretsRegex(secretRegex string) string {
	whitespaceRegex := regexp.MustCompile(`(?m)\|?\\s(\||\.|\+|\*|\{\d+,?\d{0,}})?`)
	secretRegex = whitespaceRegex.ReplaceAllString(secretRegex, "")
	return secretRegex
}

func formatSecretsNameRegex(secretNameRegex string) string {

	secretNameRegex = replaceLikeForLike(
		secretNameRegex,
		[]string{"api", "service"},
	)

	secretNameRegex = replaceLikeForLike(
		secretNameRegex,
		[]string{"secret", "token", "key"},
	)

	return secretNameRegex

}

func replaceLikeForLike(subject string, terms []string) string {
	return replaceTermsWithAlternatives(subject, terms, terms)
}

func replaceTermsWithAlternatives(subject string, terms []string, replacements []string) string {
	termsRegex := regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(terms, "|")))
	replacementsRegex := fmt.Sprintf("(%s)", strings.Join(replacements, "|"))
	return termsRegex.ReplaceAllString(subject, replacementsRegex)
}
