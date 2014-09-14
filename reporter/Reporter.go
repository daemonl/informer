package reporter

import "fmt"

type Reporter struct {
	Reports  []*Report
	Name     string
	Errors   []error
	Parent   *Reporter
	Children []*Reporter
}

type Report struct {
	Name   string
	Result string
	Warn   bool
}

func GetRoot(name string) *Reporter {
	return &Reporter{
		Name:     name,
		Reports:  []*Report{},
		Parent:   nil,
		Errors:   []error{},
		Children: []*Reporter{},
	}
}

func (r *Reporter) Spawn(name string) *Reporter {
	child := &Reporter{
		Name:     name,
		Reports:  []*Report{},
		Errors:   []error{},
		Parent:   r,
		Children: []*Reporter{},
	}
	r.Children = append(r.Children, child)
	return child
}

func (reporter *Reporter) Report(name string, params ...interface{}) *Report {
	report := &Report{
		Name: fmt.Sprintf(name, params...),
	}
	reporter.Reports = append(reporter.Reports, report)
	return report
}

func (reporter *Reporter) AddError(err error) {
	reporter.Errors = append(reporter.Errors, err)
}

func (r *Report) Fail(result string, params ...interface{}) {
	r.setResult(false, result, params...)
}
func (r *Report) Pass(result string, params ...interface{}) {
	r.setResult(true, result, params...)
}
func (r *Report) setResult(pass bool, result string, params ...interface{}) {
	r.Warn = !pass
	r.Result = fmt.Sprintf(result, params...)
}

func (r *Reporter) CollectWarnings() []string {
	return r.getWarnings(r.Name)
}

func (r *Reporter) getWarnings(name string) []string {
	w := []string{}
	for _, err := range r.Errors {
		w = append(w, fmt.Sprintf("%s ERR: %s\n", name, err.Error()))
	}

	for _, report := range r.Reports {
		if report.Warn {
			w = append(w, fmt.Sprintf("%s: %s \n %s\n", name, report.Name, report.Result))
		}
	}

	for _, c := range r.Children {
		w = append(w, c.getWarnings(name+"."+c.Name)...)
	}

	return w
}

func (r *Reporter) DumpReport() {
	r.writeReports(r.Name)

}

func (r *Reporter) writeReports(name string) {
	for _, err := range r.Errors {
		fmt.Printf("%s ERR: %s\n", name, err.Error())
	}
	for _, report := range r.Reports {
		stat := "PASS"
		if report.Warn {
			stat = "FAIL"
		}
		fmt.Printf("%s: %s \n   - %s - %s\n", name, report.Name, stat, report.Result)
	}
	for _, c := range r.Children {
		c.writeReports(name + "." + c.Name)
	}
}
