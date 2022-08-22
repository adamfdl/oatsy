package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
        "github.com/fatih/color"
)

type testResult int

const (
	testResultPassed testResult = iota
	testResultFailed
)

var testResultToString = map[testResult]string{
	testResultPassed: "PASSED",
	testResultFailed: "FAILED",
}

func (t testResult) String() string {
	return testResultToString[t]
}

type actualExpectStatusCodes struct {
        Expected int
        Actual int
}

func (a actualExpectStatusCodes) Match() bool {
    return a.Expected == a.Actual
}

type actualExpectBody struct {
        Expected string
        Actual string
}

func (a actualExpectBody) Match() bool {
    return a.Expected == a.Actual
}

type resultDetails struct {
        HTTPStatusCodes actualExpectStatusCodes
        Body actualExpectBody
}

func (rd *resultDetails) SetActualExpectHTTPStatus(expected, actual int) {
        rd.HTTPStatusCodes.Expected = expected
        rd.HTTPStatusCodes.Actual = actual
}

func (rd *resultDetails) SetActualExpectBody(expected, actual string) {
        var p1 bytes.Buffer
        err := json.Indent(&p1, []byte(expected), "", "  ")
        if err != nil {
                rd.Body.Expected = expected
        } else {
                rd.Body.Expected = p1.String()

        }


        rd.Body.Actual = actual
}

type testSuiteReport struct {
	PathName      string
	Description   string
	Operation     string
	Status        testResult
        ResultDetails resultDetails

        FailureReason string

        // This error is caused by http client returning error, we cannot assert the statusCodes
        // or bodies if there are error in the http call
        Err error
}

func (t testSuiteReport) IsPassed() bool {
    return t.Status == testResultPassed
}

func (t *testSuiteReport) Fail() {
        t.Status = testResultFailed
}

func (t *testSuiteReport) FailWithError(err error) {
        t.Err = err
        t.Fail()
}

func (t *testSuiteReport) Pass() {
	t.Status = testResultPassed
}

type Report struct {
	TestSuites []testSuiteReport
}

func (r Report) AllSuccess() bool {
        for _, testSuiteReport := range r.TestSuites {
            if !testSuiteReport.IsPassed() {
                    return false
            }
        }
        return true
}


func (r Report) generateFailingTestDescriptions() {

        fmt.Println()
        fmt.Println()

        for _, testSuiteReport := range r.TestSuites {

                if !testSuiteReport.IsPassed() {

                        color.New(color.FgRed).Add(color.Bold).Printf("• %s > %s > %s\n\n", testSuiteReport.PathName, testSuiteReport.Operation, testSuiteReport.Description)

                        if testSuiteReport.Err != nil {
                                fmt.Printf("\t Received error: %s\n\n", testSuiteReport.Err.Error())
                                continue
                        }

                        var p1 = fmt.Sprintf("Expected status code: \t%d\nActual status code: \t%d\n\n", 
                            testSuiteReport.ResultDetails.HTTPStatusCodes.Expected,
                            testSuiteReport.ResultDetails.HTTPStatusCodes.Actual)
                        if testSuiteReport.ResultDetails.HTTPStatusCodes.Match() {
                                color.Green(p1)
                        } else {
                                color.Red(p1)
                        }


                        var p2 = fmt.Sprintf("Expected body: \n\n%s\n\nActual body: \n\n%s\n", 
                            testSuiteReport.ResultDetails.Body.Expected, testSuiteReport.ResultDetails.Body.Actual)
                        if testSuiteReport.ResultDetails.Body.Match() {
                                color.Green(p2)
                        } else {
                                color.Red(p2)
                        }

                        fmt.Println()
                }
        }
}


func (r Report) generate() {

        data := [][]string{}

        for _, testSuiteReport := range r.TestSuites {
            data = append(data, []string{
                testSuiteReport.PathName,
                testSuiteReport.Operation,
                testSuiteReport.Description,
                testSuiteReport.Status.String(),
            })
        }

        table := tablewriter.NewWriter(os.Stdout)
        table.SetHeader([]string{"Path", "Operation", "Desc", "Status"})
        table.SetAutoMergeCellsByColumnIndex([]int{0})
        table.SetRowLine(true)
        for i,  row := range data {

            // Row 3 is STATUS -- Refer to SetHeader function
            if row[3] == "FAILED" {

                    // The function `Rich` also appends data to tablewriter
                    // so we do not need to manually append again
                    table.Rich(data[i], []tablewriter.Colors{
                        {},
                        {},
                        {},
                        {
                            // If test is failing, we mark the cell as RED color
                            tablewriter.Bold, tablewriter.BgRedColor,
                        },
                    })

            } else {
                    table.Rich(data[i], []tablewriter.Colors{
                        {},
                        {},
                        {},
                        {
                            // If test is failing, we mark the cell as GREEN color
                            tablewriter.Bold, tablewriter.FgGreenColor,
                        },
                    })

            }

        }

        table.Render()

        r.generateFailingTestDescriptions()
}
