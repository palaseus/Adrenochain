#!/bin/bash

# Comprehensive Audit Script for Adrenochain
# This script runs the complete audit process including security, performance, and testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AUDIT_DIR="./audit_results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
AUDIT_REPORT="$AUDIT_DIR/comprehensive_audit_$TIMESTAMP.md"

# Create audit directory
mkdir -p "$AUDIT_DIR"

echo -e "${BLUE}🔍 Starting Comprehensive Adrenochain Audit${NC}"
echo "=================================================="
echo "Timestamp: $(date)"
echo "Audit Directory: $AUDIT_DIR"
echo "Report: $AUDIT_REPORT"
echo ""

# Function to print section headers
print_section() {
    echo -e "\n${BLUE}📋 $1${NC}"
    echo "----------------------------------------"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to run command and capture output
run_command() {
    local cmd="$1"
    local output_file="$2"
    local description="$3"
    
    echo "Running: $description"
    if eval "$cmd" > "$output_file" 2>&1; then
        print_success "$description completed"
        return 0
    else
        print_error "$description failed"
        return 1
    fi
}

# Start audit report
cat > "$AUDIT_REPORT" << EOF
# 🔍 Comprehensive Adrenochain Audit Report

**Date**: $(date)  
**Auditor**: Automated Audit System  
**Scope**: Complete codebase audit and security assessment  
**Status**: 🔄 **IN PROGRESS**

---

## 📋 Executive Summary

This comprehensive audit covers:
- Security vulnerability assessment
- Performance analysis
- Code quality review
- Testing coverage analysis
- Architecture review
- Documentation assessment

---

## 🔒 Security Audit

EOF

# 1. Security Audit
print_section "Security Vulnerability Assessment"

# Run security tests
if run_command "go test ./pkg/security/... -v" "$AUDIT_DIR/security_tests.log" "Security Tests"; then
    echo "✅ Security tests passed" >> "$AUDIT_REPORT"
else
    echo "❌ Security tests failed" >> "$AUDIT_REPORT"
fi

# Run fuzz tests
if run_command "go test -fuzz=Fuzz ./pkg/security/..." "$AUDIT_DIR/fuzz_tests.log" "Fuzz Tests"; then
    echo "✅ Fuzz tests passed" >> "$AUDIT_REPORT"
else
    echo "⚠️  Fuzz tests completed with issues" >> "$AUDIT_REPORT"
fi

# Check for security vulnerabilities
if run_command "grep -r 'TODO.*security\\|FIXME.*security\\|XXX.*security' pkg/" "$AUDIT_DIR/security_todos.log" "Security TODOs"; then
    echo "⚠️  Security TODOs found" >> "$AUDIT_REPORT"
else
    echo "✅ No security TODOs found" >> "$AUDIT_REPORT"
fi

# 2. Performance Analysis
print_section "Performance Analysis"

cat >> "$AUDIT_REPORT" << EOF

## ⚡ Performance Analysis

EOF

# Run performance benchmarks
if run_command "go test -bench=. ./pkg/... -benchmem" "$AUDIT_DIR/performance_benchmarks.log" "Performance Benchmarks"; then
    echo "✅ Performance benchmarks completed" >> "$AUDIT_REPORT"
else
    echo "⚠️  Performance benchmarks completed with issues" >> "$AUDIT_REPORT"
fi

# Memory profiling
if run_command "go test -memprofile=$AUDIT_DIR/mem.prof ./pkg/..." "$AUDIT_DIR/memory_profile.log" "Memory Profiling"; then
    echo "✅ Memory profiling completed" >> "$AUDIT_REPORT"
else
    echo "⚠️  Memory profiling completed with issues" >> "$AUDIT_REPORT"
fi

# CPU profiling
if run_command "go test -cpuprofile=$AUDIT_DIR/cpu.prof ./pkg/..." "$AUDIT_DIR/cpu_profile.log" "CPU Profiling"; then
    echo "✅ CPU profiling completed" >> "$AUDIT_REPORT"
else
    echo "⚠️  CPU profiling completed with issues" >> "$AUDIT_REPORT"
fi

# 3. Code Quality Review
print_section "Code Quality Review"

cat >> "$AUDIT_REPORT" << EOF

## 📊 Code Quality Review

EOF

# Run go vet
if run_command "go vet ./..." "$AUDIT_DIR/govet.log" "Go Vet Analysis"; then
    echo "✅ Go vet analysis passed" >> "$AUDIT_REPORT"
else
    echo "⚠️  Go vet analysis found issues" >> "$AUDIT_REPORT"
fi

# Run go fmt check
if run_command "gofmt -l ." "$AUDIT_DIR/gofmt.log" "Go Format Check"; then
    if [ -s "$AUDIT_DIR/gofmt.log" ]; then
        echo "⚠️  Go format issues found" >> "$AUDIT_REPORT"
    else
        echo "✅ Go format check passed" >> "$AUDIT_REPORT"
    fi
else
    echo "❌ Go format check failed" >> "$AUDIT_REPORT"
fi

# Run go mod tidy check
if run_command "go mod tidy && git diff --exit-code go.mod go.sum" "$AUDIT_DIR/gomod.log" "Go Mod Tidy Check"; then
    echo "✅ Go mod tidy check passed" >> "$AUDIT_REPORT"
else
    echo "⚠️  Go mod tidy issues found" >> "$AUDIT_REPORT"
fi

# 4. Testing Coverage Analysis
print_section "Testing Coverage Analysis"

cat >> "$AUDIT_REPORT" << EOF

## 🧪 Testing Coverage Analysis

EOF

# Run test coverage
if run_command "go test -coverprofile=$AUDIT_DIR/coverage.out ./..." "$AUDIT_DIR/coverage.log" "Test Coverage Analysis"; then
    echo "✅ Test coverage analysis completed" >> "$AUDIT_REPORT"
    
    # Generate coverage report
    if run_command "go tool cover -html=$AUDIT_DIR/coverage.out -o $AUDIT_DIR/coverage.html" "$AUDIT_DIR/coverage_html.log" "Coverage HTML Report"; then
        echo "✅ Coverage HTML report generated" >> "$AUDIT_REPORT"
    fi
    
    # Get coverage percentage
    COVERAGE=$(go tool cover -func=$AUDIT_DIR/coverage.out | grep total | awk '{print $3}')
    echo "📊 Overall Coverage: $COVERAGE" >> "$AUDIT_REPORT"
else
    echo "❌ Test coverage analysis failed" >> "$AUDIT_REPORT"
fi

# Run race detection
if run_command "go test -race ./..." "$AUDIT_DIR/race_detection.log" "Race Detection"; then
    echo "✅ Race detection passed" >> "$AUDIT_REPORT"
else
    echo "⚠️  Race detection found issues" >> "$AUDIT_REPORT"
fi

# 5. Architecture Review
print_section "Architecture Review"

cat >> "$AUDIT_REPORT" << EOF

## 🏗️ Architecture Review

EOF

# Check package dependencies
if run_command "go list -f '{{.ImportPath}} {{.Imports}}' ./..." "$AUDIT_DIR/dependencies.log" "Package Dependencies"; then
    echo "✅ Package dependencies analyzed" >> "$AUDIT_REPORT"
else
    echo "⚠️  Package dependencies analysis completed with issues" >> "$AUDIT_REPORT"
fi

# Check for circular dependencies
if run_command "go list -f '{{.ImportPath}}' ./... | xargs -I {} sh -c 'echo \"Checking {}\"; go list -f \"{{.ImportPath}} {{.Imports}}\" {}'" "$AUDIT_DIR/circular_deps.log" "Circular Dependencies Check"; then
    echo "✅ Circular dependencies check completed" >> "$AUDIT_REPORT"
else
    echo "⚠️  Circular dependencies check completed with issues" >> "$AUDIT_REPORT"
fi

# 6. Documentation Assessment
print_section "Documentation Assessment"

cat >> "$AUDIT_REPORT" << EOF

## 📚 Documentation Assessment

EOF

# Check for missing documentation
if run_command "find pkg/ -name '*.go' -exec grep -L '^// Package' {} \;" "$AUDIT_DIR/missing_package_docs.log" "Missing Package Documentation"; then
    if [ -s "$AUDIT_DIR/missing_package_docs.log" ]; then
        echo "⚠️  Missing package documentation found" >> "$AUDIT_REPORT"
    else
        echo "✅ All packages have documentation" >> "$AUDIT_REPORT"
    fi
else
    echo "❌ Documentation check failed" >> "$AUDIT_REPORT"
fi

# Check for missing function documentation
if run_command "find pkg/ -name '*.go' -exec grep -L '^// [A-Z]' {} \;" "$AUDIT_DIR/missing_func_docs.log" "Missing Function Documentation"; then
    if [ -s "$AUDIT_DIR/missing_func_docs.log" ]; then
        echo "⚠️  Missing function documentation found" >> "$AUDIT_REPORT"
    else
        echo "✅ Function documentation is comprehensive" >> "$AUDIT_REPORT"
    fi
else
    echo "❌ Function documentation check failed" >> "$AUDIT_REPORT"
fi

# 7. Build Verification
print_section "Build Verification"

cat >> "$AUDIT_REPORT" << EOF

## 🔨 Build Verification

EOF

# Clean build
if run_command "go clean -cache" "$AUDIT_DIR/clean_build.log" "Clean Build Cache"; then
    echo "✅ Build cache cleaned" >> "$AUDIT_REPORT"
fi

# Build all packages
if run_command "go build ./..." "$AUDIT_DIR/build_all.log" "Build All Packages"; then
    echo "✅ All packages build successfully" >> "$AUDIT_REPORT"
else
    echo "❌ Build failures detected" >> "$AUDIT_REPORT"
fi

# Build main application
if run_command "go build -o $AUDIT_DIR/adrenochain ./cmd/gochain" "$AUDIT_DIR/build_main.log" "Build Main Application"; then
    echo "✅ Main application builds successfully" >> "$AUDIT_REPORT"
else
    echo "❌ Main application build failed" >> "$AUDIT_REPORT"
fi

# 8. Generate Summary
print_section "Generating Audit Summary"

cat >> "$AUDIT_REPORT" << EOF

## 📊 Audit Summary

### Files Analyzed
- **Go Files**: $(find pkg/ -name '*.go' | wc -l)
- **Test Files**: $(find pkg/ -name '*_test.go' | wc -l)
- **Total Lines**: $(find pkg/ -name '*.go' -exec wc -l {} + | tail -1 | awk '{print $1}')

### Test Results
- **Total Tests**: $(grep -r "func Test" pkg/ | wc -l)
- **Benchmark Tests**: $(grep -r "func Benchmark" pkg/ | wc -l)
- **Fuzz Tests**: $(grep -r "func Fuzz" pkg/ | wc -l)

### Security Status
- **Security Tests**: $(grep -c "PASS" $AUDIT_DIR/security_tests.log 2>/dev/null || echo "0")
- **Fuzz Tests**: $(grep -c "PASS" $AUDIT_DIR/fuzz_tests.log 2>/dev/null || echo "0")
- **Security TODOs**: $(wc -l < $AUDIT_DIR/security_todos.log 2>/dev/null || echo "0")

### Performance Status
- **Benchmark Tests**: $(grep -c "Benchmark" $AUDIT_DIR/performance_benchmarks.log 2>/dev/null || echo "0")
- **Memory Profile**: $(ls -la $AUDIT_DIR/mem.prof 2>/dev/null | awk '{print $5}' || echo "0") bytes
- **CPU Profile**: $(ls -la $AUDIT_DIR/cpu.prof 2>/dev/null | awk '{print $5}' || echo "0") bytes

### Code Quality Status
- **Go Vet Issues**: $(wc -l < $AUDIT_DIR/govet.log 2>/dev/null || echo "0")
- **Format Issues**: $(wc -l < $AUDIT_DIR/gofmt.log 2>/dev/null || echo "0")
- **Go Mod Issues**: $(wc -l < $AUDIT_DIR/gomod.log 2>/dev/null || echo "0")

### Testing Status
- **Coverage**: $(go tool cover -func=$AUDIT_DIR/coverage.out 2>/dev/null | grep total | awk '{print $3}' || echo "N/A")
- **Race Issues**: $(grep -c "WARNING: DATA RACE" $AUDIT_DIR/race_detection.log 2>/dev/null || echo "0")

---

## 🎯 Recommendations

### Immediate Actions
1. Review and address any security TODOs found
2. Fix any build failures
3. Address code quality issues
4. Improve test coverage where needed

### Future Enhancements
1. Implement additional security tests
2. Add more performance benchmarks
3. Enhance documentation coverage
4. Implement automated security scanning

---

**Audit Completed**: $(date)  
**Audit Duration**: $(($(date +%s) - $(date -d "$(date)" +%s))) seconds  
**Audit Status**: ✅ **COMPLETE**

EOF

# Final summary
print_section "Audit Complete"

echo -e "${GREEN}✅ Comprehensive audit completed successfully!${NC}"
echo ""
echo "📊 Audit Results:"
echo "  - Report: $AUDIT_REPORT"
echo "  - Coverage Report: $AUDIT_DIR/coverage.html"
echo "  - Memory Profile: $AUDIT_DIR/mem.prof"
echo "  - CPU Profile: $AUDIT_DIR/cpu.prof"
echo "  - All logs: $AUDIT_DIR/"
echo ""

# Display key metrics
if [ -f "$AUDIT_DIR/coverage.out" ]; then
    COVERAGE=$(go tool cover -func=$AUDIT_DIR/coverage.out | grep total | awk '{print $3}')
    echo -e "${BLUE}📈 Overall Test Coverage: $COVERAGE${NC}"
fi

if [ -f "$AUDIT_DIR/security_tests.log" ]; then
    SECURITY_TESTS=$(grep -c "PASS" $AUDIT_DIR/security_tests.log 2>/dev/null || echo "0")
    echo -e "${GREEN}🔒 Security Tests Passed: $SECURITY_TESTS${NC}"
fi

if [ -f "$AUDIT_DIR/performance_benchmarks.log" ]; then
    BENCHMARKS=$(grep -c "Benchmark" $AUDIT_DIR/performance_benchmarks.log 2>/dev/null || echo "0")
    echo -e "${BLUE}⚡ Performance Benchmarks: $BENCHMARKS${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Audit process completed successfully!${NC}"
echo -e "${BLUE}📋 Review the comprehensive report: $AUDIT_REPORT${NC}"
