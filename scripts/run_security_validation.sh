#!/bin/bash

echo "🔒 Adrenochain Comprehensive Security Validation Suite"
echo "======================================================"
echo ""

# Build the security validation tool
echo "🔨 Building security validation tool..."
go build -o security_tool ./cmd/security

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"

# Run the security validation
echo ""
echo "🔒 Running comprehensive security validation..."
./security_tool

if [ $? -ne 0 ]; then
    echo "❌ Security validation failed!"
    exit 1
fi

echo ""
echo "🎉 All security validations completed successfully!"
echo "📄 Check the generated JSON report for detailed results."

# Clean up
rm -f security_tool

echo ""
echo "✨ Security validation session completed!"
