#!/bin/bash

echo "🚀 Adrenochain Comprehensive Performance Benchmarking Suite"
echo "============================================================"
echo ""

# Build the benchmark tool
echo "🔨 Building benchmark tool..."
go build -o benchmark_tool ./cmd/benchmark

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Build successful!"

# Run the benchmarks
echo ""
echo "📊 Running comprehensive benchmarks..."
./benchmark_tool

if [ $? -ne 0 ]; then
    echo "❌ Benchmarking failed!"
    exit 1
fi

echo ""
echo "🎉 All benchmarks completed successfully!"
echo "📄 Check the generated JSON report for detailed results."

# Clean up
rm -f benchmark_tool

echo ""
echo "✨ Benchmarking session completed!"
