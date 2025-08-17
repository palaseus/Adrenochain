#!/usr/bin/env python3
"""
adrenochain Test Analyzer
Provides advanced analysis of test results, performance metrics, and trend tracking.
"""

import os
import sys
import json
import argparse
import subprocess
import re
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Any, Optional
import matplotlib.pyplot as plt
import pandas as pd

class TestAnalyzer:
    def __init__(self, project_root: str):
        self.project_root = Path(project_root)
        self.test_results_dir = self.project_root / "test_results"
        self.coverage_dir = self.project_root / "coverage"
        self.analysis_dir = self.project_root / "test_analysis"
        
        # Create analysis directory
        self.analysis_dir.mkdir(exist_ok=True)
        
    def analyze_test_results(self) -> Dict[str, Any]:
        """Analyze all test result files and generate comprehensive report."""
        print("🔍 Analyzing test results...")
        
        analysis = {
            "timestamp": datetime.now().isoformat(),
            "summary": {},
            "packages": {},
            "performance": {},
            "coverage": {},
            "trends": {}
        }
        
        # Analyze test results
        if self.test_results_dir.exists():
            analysis["packages"] = self._analyze_package_results()
            analysis["performance"] = self._analyze_performance()
        
        # Analyze coverage
        if self.coverage_dir.exists():
            analysis["coverage"] = self._analyze_coverage()
        
        # Generate summary
        analysis["summary"] = self._generate_summary(analysis)
        
        # Save analysis
        analysis_file = self.analysis_dir / f"analysis_{datetime.now().strftime('%Y%m%d_%H%M%S')}.json"
        with open(analysis_file, 'w') as f:
            json.dump(analysis, f, indent=2)
        
        print(f"✅ Analysis saved to: {analysis_file}")
        return analysis
    
    def _analyze_package_results(self) -> Dict[str, Any]:
        """Analyze individual package test results."""
        packages = {}
        
        for log_file in self.test_results_dir.glob("*_tests.log"):
            package_name = log_file.stem.replace("_tests", "")
            
            with open(log_file, 'r') as f:
                content = f.read()
            
            # Parse test results
            test_count = len(re.findall(r"=== RUN", content))
            passed_count = len(re.findall(r"--- PASS", content))
            failed_count = len(re.findall(r"--- FAIL", content))
            skipped_count = len(re.findall(r"--- SKIP", content))
            
            # Extract timing information
            timing_match = re.search(r"(\d+\.?\d*)s", content)
            duration = float(timing_match.group(1)) if timing_match else 0
            
            packages[package_name] = {
                "test_count": test_count,
                "passed": passed_count,
                "failed": failed_count,
                "skipped": skipped_count,
                "duration": duration,
                "success_rate": (passed_count / test_count * 100) if test_count > 0 else 0
            }
        
        return packages
    
    def _analyze_performance(self) -> Dict[str, Any]:
        """Analyze benchmark and performance results."""
        performance = {}
        
        # Analyze benchmark results
        for bench_file in self.test_results_dir.glob("*_bench.log"):
            package_name = bench_file.stem.replace("_bench", "")
            
            with open(bench_file, 'r') as f:
                content = f.read()
            
            # Parse benchmark results
            benchmarks = []
            for line in content.split('\n'):
                if line.startswith('Benchmark'):
                    parts = line.split()
                    if len(parts) >= 4:
                        name = parts[0]
                        iterations = int(parts[1])
                        time_per_op = float(parts[2])
                        memory_per_op = parts[3] if len(parts) > 3 else "N/A"
                        
                        benchmarks.append({
                            "name": name,
                            "iterations": iterations,
                            "time_per_op": time_per_op,
                            "memory_per_op": memory_per_op
                        })
            
            performance[package_name] = {
                "benchmarks": benchmarks,
                "benchmark_count": len(benchmarks)
            }
        
        return performance
    
    def _analyze_coverage(self) -> Dict[str, Any]:
        """Analyze code coverage data."""
        coverage = {}
        
        # Find coverage files
        coverage_files = list(self.coverage_dir.glob("*_coverage.out"))
        
        if not coverage_files:
            return coverage
        
        # Combine coverage data
        combined_file = self.coverage_dir / "combined_coverage.out"
        if len(coverage_files) > 1:
            # Combine multiple coverage files
            with open(combined_file, 'w') as outfile:
                outfile.write("mode: atomic\n")
                for file in coverage_files:
                    with open(file, 'r') as infile:
                        for line in infile:
                            if not line.startswith("mode:"):
                                outfile.write(line)
        else:
            combined_file = coverage_files[0]
        
        # Parse coverage data
        if combined_file.exists():
            try:
                result = subprocess.run(
                    ["go", "tool", "cover", "-func", str(combined_file)],
                    capture_output=True, text=True, cwd=self.project_root
                )
                
                if result.returncode == 0:
                    lines = result.stdout.split('\n')
                    total_coverage = 0
                    
                    for line in lines:
                        if line.startswith("total:"):
                            coverage_match = re.search(r"(\d+\.?\d*)%", line)
                            if coverage_match:
                                total_coverage = float(coverage_match.group(1))
                            break
                    
                    coverage["total"] = total_coverage
                    coverage["combined_file"] = str(combined_file)
            except Exception as e:
                coverage["error"] = str(e)
        
        return coverage
    
    def _generate_summary(self, analysis: Dict[str, Any]) -> Dict[str, Any]:
        """Generate summary statistics from analysis."""
        summary = {}
        
        # Package summary
        packages = analysis.get("packages", {})
        if packages:
            total_packages = len(packages)
            total_tests = sum(pkg.get("test_count", 0) for pkg in packages.values())
            total_passed = sum(pkg.get("passed", 0) for pkg in packages.values())
            total_failed = sum(pkg.get("failed", 0) for pkg in packages.values())
            total_skipped = sum(pkg.get("skipped", 0) for pkg in packages.values())
            total_duration = sum(pkg.get("duration", 0) for pkg in packages.values())
            
            summary["packages"] = {
                "total": total_packages,
                "tests": {
                    "total": total_tests,
                    "passed": total_passed,
                    "failed": total_failed,
                    "skipped": total_skipped
                },
                "duration": total_duration,
                "success_rate": (total_passed / total_tests * 100) if total_tests > 0 else 0
            }
        
        # Performance summary
        performance = analysis.get("performance", {})
        if performance:
            total_benchmarks = sum(pkg.get("benchmark_count", 0) for pkg in performance.values())
            summary["performance"] = {
                "total_benchmarks": total_benchmarks,
                "packages_with_benchmarks": len(performance)
            }
        
        # Coverage summary
        coverage = analysis.get("coverage", {})
        if coverage and "total" in coverage:
            summary["coverage"] = {
                "total": coverage["total"]
            }
        
        return summary
    
    def generate_visualizations(self, analysis: Dict[str, Any]):
        """Generate visual charts and graphs from analysis data."""
        print("📊 Generating visualizations...")
        
        try:
            # Test results pie chart
            packages = analysis.get("packages", {})
            if packages:
                self._create_test_results_chart(packages)
                self._create_performance_chart(packages)
            
            # Coverage chart
            coverage = analysis.get("coverage", {})
            if coverage and "total" in coverage:
                self._create_coverage_chart(coverage)
            
            print("✅ Visualizations generated successfully")
            
        except ImportError:
            print("⚠️  matplotlib not available. Install with: pip install matplotlib")
        except Exception as e:
            print(f"❌ Error generating visualizations: {e}")
    
    def _create_test_results_chart(self, packages: Dict[str, Any]):
        """Create test results visualization."""
        fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(15, 6))
        
        # Package success rates
        names = list(packages.keys())
        success_rates = [pkg.get("success_rate", 0) for pkg in packages.values()]
        
        ax1.bar(names, success_rates, color=['green' if rate == 100 else 'orange' if rate >= 80 else 'red' for rate in success_rates])
        ax1.set_title("Package Test Success Rates")
        ax1.set_ylabel("Success Rate (%)")
        ax1.tick_params(axis='x', rotation=45)
        
        # Test distribution
        test_counts = [pkg.get("test_count", 0) for pkg in packages.values()]
        ax2.pie(test_counts, labels=names, autopct='%1.1f%%')
        ax2.set_title("Test Distribution by Package")
        
        plt.tight_layout()
        plt.savefig(self.analysis_dir / "test_results_chart.png", dpi=300, bbox_inches='tight')
        plt.close()
    
    def _create_performance_chart(self, packages: Dict[str, Any]):
        """Create performance visualization."""
        fig, ax = plt.subplots(figsize=(12, 6))
        
        names = list(packages.keys())
        durations = [pkg.get("duration", 0) for pkg in packages.values()]
        
        bars = ax.bar(names, durations, color='skyblue')
        ax.set_title("Test Execution Time by Package")
        ax.set_ylabel("Duration (seconds)")
        ax.tick_params(axis='x', rotation=45)
        
        # Add value labels on bars
        for bar, duration in zip(bars, durations):
            height = bar.get_height()
            ax.text(bar.get_x() + bar.get_width()/2., height,
                   f'{duration:.1f}s', ha='center', va='bottom')
        
        plt.tight_layout()
        plt.savefig(self.analysis_dir / "performance_chart.png", dpi=300, bbox_inches='tight')
        plt.close()
    
    def _create_coverage_chart(self, coverage: Dict[str, Any]):
        """Create coverage visualization."""
        fig, ax = plt.subplots(figsize=(8, 6))
        
        total = coverage.get("total", 0)
        remaining = 100 - total
        
        labels = ['Covered', 'Uncovered']
        sizes = [total, remaining]
        colors = ['lightgreen', 'lightcoral']
        
        ax.pie(sizes, labels=labels, colors=colors, autopct='%1.1f%%', startangle=90)
        ax.set_title(f"Code Coverage: {total:.1f}%")
        
        plt.tight_layout()
        plt.savefig(self.analysis_dir / "coverage_chart.png", dpi=300, bbox_inches='tight')
        plt.close()
    
    def generate_html_report(self, analysis: Dict[str, Any]):
        """Generate an HTML report from the analysis."""
        print("📄 Generating HTML report...")
        
        html_content = f"""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>adrenochain Test Analysis Report</title>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }}
        .container {{ max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }}
        .header {{ text-align: center; border-bottom: 2px solid #007acc; padding-bottom: 20px; margin-bottom: 30px; }}
        .metric {{ background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #007acc; }}
        .package {{ background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; }}
        .success {{ border-left-color: #28a745; }}
        .warning {{ border-left-color: #ffc107; }}
        .danger {{ border-left-color: #dc3545; }}
        .chart {{ text-align: center; margin: 20px 0; }}
        table {{ width: 100%; border-collapse: collapse; margin: 20px 0; }}
        th, td {{ padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }}
        th {{ background-color: #007acc; color: white; }}
        .timestamp {{ color: #666; font-size: 0.9em; }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 adrenochain Test Analysis Report</h1>
            <p class="timestamp">Generated: {analysis.get('timestamp', 'Unknown')}</p>
        </div>
        
        <div class="metric">
            <h2>📊 Summary</h2>
            <p><strong>Total Packages:</strong> {analysis.get('summary', {}).get('packages', {}).get('total', 0)}</p>
            <p><strong>Total Tests:</strong> {analysis.get('summary', {}).get('packages', {}).get('tests', {}).get('total', 0)}</p>
            <p><strong>Success Rate:</strong> {analysis.get('summary', {}).get('packages', {}).get('success_rate', 0):.1f}%</p>
            <p><strong>Total Duration:</strong> {analysis.get('summary', {}).get('packages', {}).get('duration', 0):.1f}s</p>
        </div>
        
        <div class="metric">
            <h2>📦 Package Details</h2>
            <table>
                <thead>
                    <tr>
                        <th>Package</th>
                        <th>Tests</th>
                        <th>Passed</th>
                        <th>Failed</th>
                        <th>Skipped</th>
                        <th>Duration</th>
                        <th>Success Rate</th>
                    </tr>
                </thead>
                <tbody>
"""
        
        # Add package rows
        for pkg_name, pkg_data in analysis.get('packages', {}).items():
            success_rate = pkg_data.get('success_rate', 0)
            row_class = 'success' if success_rate == 100 else 'warning' if success_rate >= 80 else 'danger'
            
            html_content += f"""
                    <tr class="{row_class}">
                        <td><strong>{pkg_name}</strong></td>
                        <td>{pkg_data.get('test_count', 0)}</td>
                        <td>{pkg_data.get('passed', 0)}</td>
                        <td>{pkg_data.get('failed', 0)}</td>
                        <td>{pkg_data.get('skipped', 0)}</td>
                        <td>{pkg_data.get('duration', 0):.1f}s</td>
                        <td>{success_rate:.1f}%</td>
                    </tr>
"""
        
        html_content += """
                </tbody>
            </table>
        </div>
        
        <div class="metric">
            <h2>📊 Coverage</h2>
"""
        
        coverage = analysis.get('coverage', {})
        if 'total' in coverage:
            html_content += f"""
            <p><strong>Total Coverage:</strong> {coverage['total']:.1f}%</p>
            <div class="chart">
                <img src="coverage_chart.png" alt="Coverage Chart" style="max-width: 100%; height: auto;">
            </div>
"""
        else:
            html_content += "<p>No coverage data available</p>"
        
        html_content += """
        </div>
        
        <div class="metric">
            <h2>📈 Performance</h2>
            <div class="chart">
                <img src="performance_chart.png" alt="Performance Chart" style="max-width: 100%; height: auto;">
            </div>
        </div>
        
        <div class="metric">
            <h2>🎯 Recommendations</h2>
"""
        
        # Generate recommendations
        summary = analysis.get('summary', {})
        packages_data = analysis.get('packages', {})
        
        if packages_data:
            failed_packages = [name for name, data in packages_data.items() if data.get('failed', 0) > 0]
            slow_packages = [name for name, data in packages_data.items() if data.get('duration', 0) > 10]
            
            if failed_packages:
                html_content += f"<p>❌ <strong>Fix failing tests in:</strong> {', '.join(failed_packages)}</p>"
            
            if slow_packages:
                html_content += f"<p>⏱️ <strong>Optimize slow tests in:</strong> {', '.join(slow_packages)}</p>"
            
            coverage_data = analysis.get('coverage', {})
            if 'total' in coverage_data and coverage_data['total'] < 80:
                html_content += f"<p>📊 <strong>Improve code coverage:</strong> Current coverage is {coverage_data['total']:.1f}%, aim for 80%+</p>"
        
        html_content += """
        </div>
    </div>
</body>
</html>
"""
        
        # Save HTML report
        html_file = self.analysis_dir / "test_analysis_report.html"
        with open(html_file, 'w') as f:
            f.write(html_content)
        
        print(f"✅ HTML report generated: {html_file}")

def main():
    parser = argparse.ArgumentParser(description="adrenochain Test Analyzer")
    parser.add_argument("--project-root", default=".", help="Project root directory")
    parser.add_argument("--visualize", action="store_true", help="Generate visualizations")
    parser.add_argument("--html", action="store_true", help="Generate HTML report")
    
    args = parser.parse_args()
    
    # Initialize analyzer
    analyzer = TestAnalyzer(args.project_root)
    
    # Run analysis
    analysis = analyzer.analyze_test_results()
    
    # Print summary
    print("\n📊 Analysis Summary:")
    summary = analysis.get("summary", {})
    if "packages" in summary:
        pkg_summary = summary["packages"]
        print(f"   📦 Packages: {pkg_summary.get('total', 0)}")
        print(f"   🧪 Tests: {pkg_summary.get('tests', {}).get('total', 0)}")
        print(f"   ✅ Success Rate: {pkg_summary.get('success_rate', 0):.1f}%")
        print(f"   ⏱️  Duration: {pkg_summary.get('duration', 0):.1f}s")
    
    if "coverage" in summary:
        print(f"   📊 Coverage: {summary['coverage'].get('total', 0):.1f}%")
    
    # Generate additional outputs
    if args.visualize:
        analyzer.generate_visualizations(analysis)
    
    if args.html:
        analyzer.generate_html_report(analysis)
    
    print(f"\n📁 Analysis files saved to: {analyzer.analysis_dir}")

if __name__ == "__main__":
    main()
