package handler

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

type FactProcessor struct {
	Fact          LagoonFact
	InsightsData  interface{}
	hasErrorState bool  //This is a private field used for the ParserFilter interface implementation
	theError      error //This is a private field used for the ParserFilter interface implementation
	filteredOut   bool  //This is set to true if a filter fails, no further processing will occur on this Fact
}

func (i *FactProcessor) getFact() LagoonFact {
	return i.Fact
}

func (i *FactProcessor) isFilteredOut() bool {
	return i.filteredOut
}

func (i *FactProcessor) setFactField(fieldname string, value string) parserFilter {
	if i.hasError() || i.isFilteredOut() {
		return i
	}

	rv := reflect.Indirect(reflect.ValueOf(&i.Fact))
	f := rv.FieldByName(fieldname)
	if !f.IsValid() {
		i.hasErrorState = true
		i.theError = errors.New("Could not find fieldname in LagoonFact: " + fieldname)
		return i
	}
	if f.Kind() != reflect.String {
		i.hasErrorState = true
		i.theError = errors.New("Can not set non-string fields on LagoonFact: " + fieldname)
		return i
	}

	f.SetString(value)

	return i
}

func (i *FactProcessor) hasError() bool {
	return i.hasErrorState
}

func (i *FactProcessor) setKeyFact(isKeyFact bool) parserFilter {
	if i.hasError() || i.isFilteredOut() {
		return i
	}
	i.Fact.KeyFact = isKeyFact
	return i
}

func (i *FactProcessor) fieldContains(fieldname string, regex string) parserFilter {
	if i.hasError() || i.isFilteredOut() {
		return i
	}

	rv := reflect.ValueOf(i.InsightsData)
	f := rv.FieldByName(fieldname)
	if !f.IsValid() {
		i.hasErrorState = true
		i.theError = errors.New("Could not find fieldname: " + fieldname)
		return i
	}
	if f.Kind() != reflect.String {
		i.hasErrorState = true
		i.theError = errors.New("Can not match regex on non-string for field: " + fieldname)
		return i
	}

	hasMatch, err := regexp.Match(regex, []byte(f.String()))
	if err != nil {
		i.hasErrorState = true
		i.theError = err
		return i
	}

	if !hasMatch {
		i.hasErrorState = true
		i.theError = errors.New("Can not match regex '" + regex + "' for field: " + fieldname)
		return i
	}

	return i
}

func (i *FactProcessor) getError() error {
	return i.theError
}

func (i *FactProcessor) isOfType(typename string) parserFilter {
	if i.hasError() || i.isFilteredOut() {
		return i
	}

	formattedTypename := "handler." + typename

	if formattedTypename != fmt.Sprintf("%T", i.InsightsData) {
		i.filteredOut = true //This type doesn't match, so we filter it out.
	}
	return i
}
