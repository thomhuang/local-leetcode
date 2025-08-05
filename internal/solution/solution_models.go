package solution

type InterpretSolutionRequest struct {
	Language   string `json:"lang"`
	QuestionId string `json:"question_id"`
	TypedCode  string `json:"typed_code"`
	DataInput  string `json:"data_input"`
}

type InterpretSolutionResponse struct {
	InterpretId string `json:"interpret_id"`
	TestCase    string `json:"test_case"`
}

type CheckSolutionResponse struct {
	State                  string   `json:"state"`
	StatusCode             int      `json:"status_code"`
	Language               string   `json:"lang"`
	RunSuccessful          bool     `json:"run_success"`
	CompileError           string   `json:"compile_error"`
	StatusRuntime          string   `json:"status_runtime"`
	Memory                 int      `json:"memory"`
	DisplayRuntime         string   `json:"display_runtime"`
	CodeAnswer             []string `json:"code_answer"`
	CodeOutput             []string `json:"code_output"`
	StdOutputList          []string `json:"std_output_list"`
	ElapsedTime            int      `json:"elapsed_time"`
	TaskFinishTime         int64    `json:"task_finish_time"`
	TaskName               string   `json:"task_name"`
	ExpectedStatusCode     int      `json:"expected_status_code"`
	ExpectedLanguage       string   `json:"expected_lang"`
	ExpectedRunSuccessful  bool     `json:"expected_run_success"`
	ExpectedStatusRuntime  string   `json:"expected_status_runtime"`
	ExpectedMemory         int      `json:"expected_memory"`
	ExpectedDisplayRuntime string   `json:"expected_display_runtime"`
	ExpectedCodeAnswer     []string `json:"expected_code_answer"`
	ExpectedCodeOutput     []string `json:"expected_code_output"`
	ExpectedStdOutputList  []string `json:"expected_std_output_list"`
	ExpectedElapsedTime    int      `json:"expected_elapsed_time"`
	ExpectedTaskFinishTime int64    `json:"expected_task_finish_time"`
	ExpectedTaskName       string   `json:"expected_task_name"`
	CorrectAnswer          bool     `json:"correct_answer"`
	CompareResult          string   `json:"compare_result"`
	TotalCorrect           int      `json:"total_correct"`
	TotalTestcases         int      `json:"total_testcases"`
	RuntimePercentile      *float64 `json:"runtime_percentile"`
	StatusMemory           string   `json:"status_memory"`
	MemoryPercentile       *float64 `json:"memory_percentile"`
	PrettyLanguage         string   `json:"pretty_lang"`
	SubmissionId           string   `json:"submission_id"`
	StatusMessage          string   `json:"status_msg"`
}
