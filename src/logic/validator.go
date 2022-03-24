package logic

import (
	"fmt"
	"regexp"
	"stock-simulator-serverless/src/models"
	"strings"
)

var (
	//alphaNumeric          = regexp.MustCompile(`^[a-zA-Z1-9]+$`).MatchString
	alphaNumericDash      = regexp.MustCompile(`^[a-zA-Z1-9\-]+$`).MatchString
	alphaNumericWithSpace = regexp.MustCompile(`^[a-zA-Z1-9\s]+$`).MatchString
	sentanceValidator     = regexp.MustCompile(`[a-zA-Z1-9\.\!\s\-]`).MatchString
	upperCaseLetters      = regexp.MustCompile(`^[A-Z]+$`).MatchString
	email                 = regexp.MustCompile("(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])").MatchString
)

func validate(field, inputString string, min, max int, validator func(string) bool) error {
	lenS := len(inputString)
	if lenS != len(strings.TrimSpace(inputString)) {
		return fmt.Errorf("invalid [%s]: contains leading or trailing space", field)
	}
	if lenS < min {
		return fmt.Errorf("invalid [%s]: must be grater than %d", field, min-1)
	}
	if lenS > max {
		return fmt.Errorf("invalid [%s]: must be less than %d", field, max+1)
	}
	if validator != nil {
		if !validator(inputString) {
			return fmt.Errorf("invalid [%s]: contains invalid characters (%s)", field, inputString)
		}
	}
	return nil
}

func validateCompany(company *models.CompanyStruct) error {
	if err := validCompanyName(company.Name); err != nil {
		return err
	}
	if err := validDescription(company.Description); err != nil {
		return err
	}
	if err := validSymbol(company.Symbol); err != nil {
		return err
	}
	return nil
}

func validateUser(user *models.PrivateUserStruct) error {
	// User Validation
	if err := validLogin(user.Login); err != nil {
		return err
	}
	if err := validDescription(user.User.Description); err != nil {
		return err
	}
	if err := validDisplayName(user.User.Name); err != nil {
		return err
	}
	if err := validPassword(user.Password); err != nil {
		return err
	}
	if err := validInvestorType(user.User.InvestorType); err != nil {
		return err
	}
	if user.Email != "" {
		return validEmail(user.Email)
	}
	return nil
}

func validPassword(input string) error {
	return validate("password", input, 6, 100, nil)
}

func validEmail(input string) error {
	return validate("email", input, 3, 100, email)
}

func validDisplayName(input string) error {
	return validate("displayName", input, 4, 20, alphaNumericWithSpace)
}

func validLogin(input string) error {
	return validate("login", input, 4, 20, alphaNumericDash)
}

func validDescription(input string) error {
	return validate("description", input, 0, 250, nil)
}

func validInvestorType(input string) error {
	return validate("investor type", input, 0, 40, nil)
}

func validCompanyName(input string) error {
	return validate("company name", input, 5, 30, sentanceValidator)
}

func validSymbol(input string) error {
	return validate("symbol", input, 2, 6, upperCaseLetters)
}
