package commands_test

import (
"os"
"os/exec"
"path/filepath"
"strings"
"testing"
)

func buildBinary(t *testing.T) string {
t.Helper()
dir := t.TempDir()
bin := filepath.Join(dir, "calconv")
cmd := exec.Command("go", "build", "-o", bin, "../../../cmd/calconv")
cmd.Dir = filepath.Join(dir)
// Use the project root
projRoot, _ := filepath.Abs("../../..")
cmd.Dir = projRoot
cmd.Args = []string{"go", "build", "-o", bin, "./cmd/calconv"}
out, err := cmd.CombinedOutput()
if err != nil {
t.Fatalf("build failed: %v\n%s", err, out)
}
return bin
}

func writeTestICS(t *testing.T, dir string) string {
t.Helper()
f := filepath.Join(dir, "test.ics")
content := `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Test//Test//EN
BEGIN:VEVENT
UID:test-event-1@test
DTSTART:20240115T100000Z
DTEND:20240115T110000Z
DTSTAMP:20240115T100000Z
SUMMARY:Test Meeting
DESCRIPTION:A test event
END:VEVENT
END:VCALENDAR
`
if err := os.WriteFile(f, []byte(content), 0644); err != nil {
t.Fatal(err)
}
return f
}

func TestConvertHappyPath(t *testing.T) {
bin := buildBinary(t)
dir := t.TempDir()
input := writeTestICS(t, dir)
output := filepath.Join(dir, "output.ics")

cmd := exec.Command(bin, "convert", input, output)
out, err := cmd.CombinedOutput()
if err != nil {
t.Fatalf("convert failed: %v\n%s", err, out)
}

data, err := os.ReadFile(output)
if err != nil {
t.Fatal(err)
}
if !strings.Contains(string(data), "Test Meeting") {
t.Error("output should contain 'Test Meeting'")
}
}

func TestFormatAutoDetection(t *testing.T) {
bin := buildBinary(t)
dir := t.TempDir()
input := writeTestICS(t, dir)
output := filepath.Join(dir, "gcal-output.csv")

cmd := exec.Command(bin, "convert", input, output, "--to", "gcal")
out, err := cmd.CombinedOutput()
if err != nil {
t.Fatalf("convert failed: %v\n%s", err, out)
}

outStr := string(out)
if !strings.Contains(outStr, "Detected source format: ics") {
t.Errorf("expected format detection message in stderr, got: %s", outStr)
}
}

func TestDryRunOutput(t *testing.T) {
bin := buildBinary(t)
dir := t.TempDir()
input := writeTestICS(t, dir)
output := filepath.Join(dir, "output.ics")

cmd := exec.Command(bin, "convert", "--dry-run", input, output)
out, err := cmd.CombinedOutput()
if err != nil {
t.Fatalf("dry-run failed: %v\n%s", err, out)
}

outStr := string(out)
if !strings.Contains(outStr, "Test Meeting") {
t.Error("dry-run should list items")
}

// Output file should NOT exist after dry-run
if _, err := os.Stat(output); err == nil {
t.Error("dry-run should not create output file")
}
}

func TestPipedIO(t *testing.T) {
bin := buildBinary(t)
dir := t.TempDir()
input := writeTestICS(t, dir)

data, _ := os.ReadFile(input)
cmd := exec.Command(bin, "convert", "--from", "ics", "--to", "ics", "-", "-")
cmd.Stdin = strings.NewReader(string(data))
out, err := cmd.CombinedOutput()
if err != nil {
t.Fatalf("piped convert failed: %v\n%s", err, out)
}
if !strings.Contains(string(out), "Test Meeting") {
t.Error("piped output should contain event")
}
}

func TestInvalidInputError(t *testing.T) {
bin := buildBinary(t)
cmd := exec.Command(bin, "convert", "/nonexistent/file.ics", "/tmp/out.ics")
out, err := cmd.CombinedOutput()
if err == nil {
t.Fatal("expected error for nonexistent input")
}
if !strings.Contains(string(out), "failed to read input") {
t.Errorf("expected read error message, got: %s", out)
}
}

func TestListFormats(t *testing.T) {
bin := buildBinary(t)
cmd := exec.Command(bin, "list-formats")
out, err := cmd.CombinedOutput()
if err != nil {
t.Fatalf("list-formats failed: %v\n%s", err, out)
}
outStr := string(out)
if !strings.Contains(outStr, "ics") || !strings.Contains(outStr, "ticktick") {
t.Error("list-formats should show supported formats")
}
}

func TestValidateCommand(t *testing.T) {
bin := buildBinary(t)
dir := t.TempDir()
input := writeTestICS(t, dir)

cmd := exec.Command(bin, "validate", input)
out, err := cmd.CombinedOutput()
if err != nil {
t.Fatalf("validate failed: %v\n%s", err, out)
}
if !strings.Contains(string(out), "Items:") {
t.Errorf("validate should show item count, got: %s", out)
}
}

func TestVersionFlag(t *testing.T) {
bin := buildBinary(t)
cmd := exec.Command(bin, "--version")
out, err := cmd.CombinedOutput()
if err != nil {
t.Fatalf("version failed: %v\n%s", err, out)
}
if !strings.Contains(string(out), "calconv version") {
t.Errorf("expected version output, got: %s", out)
}
}
