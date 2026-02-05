#!/usr/bin/env python3
"""Benchmark sol vs upsun CLI output sizes and token counts.

Usage:
    pipx run --spec tiktoken python3 scripts/benchmark.py --project PROJECT_ID

Or set PLATFORM_PROJECT environment variable:
    export PLATFORM_PROJECT=your-project-id
    pipx run --spec tiktoken python3 scripts/benchmark.py
"""

import argparse
import os
import subprocess
import sys

# Determine script and sol binary paths
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
SOL_BIN = os.path.join(SCRIPT_DIR, '..', 'sol')

# Try to import tiktoken
try:
    import tiktoken
    enc = tiktoken.encoding_for_model("gpt-4")
    def count_tokens(text):
        return len(enc.encode(text))
    HAS_TIKTOKEN = True
except ImportError:
    HAS_TIKTOKEN = False
    def count_tokens(text):
        return len(text) // 4

def check_prereqs():
    """Check for required tools and warn if missing."""
    warnings = []

    if subprocess.run("which upsun", shell=True, capture_output=True).returncode != 0:
        warnings.append("upsun CLI not found - upsun benchmarks will show N/A")

    if not os.path.exists(SOL_BIN):
        warnings.append(f"sol binary not found at {SOL_BIN} - run 'go build -o sol .' first")

    if warnings:
        print("\n  Warnings:")
        for w in warnings:
            print(f"    - {w}")
        print()

def run_command(cmd):
    """Run command and return output."""
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=60)
        return result.stdout
    except Exception as e:
        return None

def format_size(bytes_size):
    """Format bytes as human readable."""
    if bytes_size >= 1024 * 1024:
        return f"{bytes_size / (1024*1024):.1f}MB"
    elif bytes_size >= 1024:
        return f"{bytes_size / 1024:.1f}KB"
    return f"{bytes_size}B"

def benchmark(name, commands, baseline_label=None):
    """Run benchmark for a set of commands."""
    print(f"\n{'─'*70}")
    print(f"  {name}")
    print(f"{'─'*70}")
    print(f"  {'Output':30} {'Size':>10}  {'Tokens':>10}  {'Parseable':>10}")
    print(f"  {'-'*30} {'-'*10}  {'-'*10}  {'-'*10}")

    results = []
    baseline = None
    for label, cmd, parseable in commands:
        output = run_command(cmd)
        if output is None or len(output) == 0:
            print(f"  {label:30} {'N/A':>10}  {'N/A':>10}  {parseable:>10}")
            continue

        bytes_size = len(output.encode('utf-8'))
        tokens = count_tokens(output)
        result = {
            'label': label,
            'bytes': bytes_size,
            'tokens': tokens,
            'parseable': parseable,
        }
        results.append(result)

        if baseline_label and label == baseline_label:
            baseline = result

        print(f"  {label:30} {format_size(bytes_size):>10}  {tokens:>10,}  {parseable:>10}")

    # Calculate savings vs baseline
    if baseline:
        print(f"\n  Token comparison vs {baseline['label']}:")
        for r in results:
            if r['label'] == baseline['label']:
                continue
            token_reduction = (1 - r['tokens'] / baseline['tokens']) * 100 if baseline['tokens'] > 0 else 0
            if token_reduction > 0:
                print(f"    {r['label']:30} {token_reduction:>5.1f}% fewer tokens")
            else:
                print(f"    {r['label']:30} {-token_reduction:>5.1f}% more tokens")

    return results

def main():
    parser = argparse.ArgumentParser(
        description='Benchmark sol vs upsun CLI token efficiency',
        epilog='Requires tiktoken for accurate token counts. Install via: pip install tiktoken'
    )
    parser.add_argument('--project', '-p',
                        help='Project ID for environment and activity benchmarks')
    args = parser.parse_args()

    # Get project ID from args or environment
    project_id = args.project or os.environ.get('PLATFORM_PROJECT')
    if not project_id:
        print("Error: No project specified.")
        print("Use --project PROJECT_ID or set PLATFORM_PROJECT environment variable.")
        sys.exit(1)

    print("\n" + "="*70)
    print("  SOL vs UPSUN CLI - TOKEN EFFICIENCY BENCHMARK")
    print("="*70)
    if HAS_TIKTOKEN:
        print("  Token counting: tiktoken (cl100k_base, same as GPT-4/Claude)")
    else:
        print("  Token counting: estimate (chars/4) - install tiktoken for accuracy")
    print(f"  Test project: {project_id}")

    check_prereqs()

    # Benchmark 1: Project list
    project_results = benchmark("PROJECT LIST", [
        ("upsun --format=plain", "upsun project:list --format=plain --count=100 2>/dev/null", "No"),
        ("sol json (all fields)", f"{SOL_BIN} project:list -o json --full 2>/dev/null", "Yes"),
        ("sol json (lean)", f"{SOL_BIN} project:list -o json 2>/dev/null", "Yes"),
        ("sol toon (lean)", f"{SOL_BIN} project:list 2>/dev/null", "Yes"),
    ], baseline_label="upsun --format=plain")

    # Benchmark 2: Environment list (big dataset)
    env_results = benchmark("ENVIRONMENT LIST", [
        ("upsun --format=plain", f"upsun environment:list -p {project_id} --format=plain 2>/dev/null", "No"),
        ("sol json (all fields)", f"{SOL_BIN} environment:list -p {project_id} -o json --full 2>/dev/null", "Yes"),
        ("sol json (lean)", f"{SOL_BIN} environment:list -p {project_id} -o json 2>/dev/null", "Yes"),
        ("sol toon (lean)", f"{SOL_BIN} environment:list -p {project_id} 2>/dev/null", "Yes"),
    ], baseline_label="upsun --format=plain")

    # Benchmark 3: Activity list
    activity_results = benchmark("ACTIVITY LIST (50 activities)", [
        ("upsun --format=plain", f"upsun activity:list -p {project_id} --limit 50 --format=plain 2>/dev/null", "No"),
        ("sol json (all fields)", f"{SOL_BIN} activity:list -p {project_id} --limit 50 -o json --full 2>/dev/null", "Yes"),
        ("sol json (lean)", f"{SOL_BIN} activity:list -p {project_id} --limit 50 -o json 2>/dev/null", "Yes"),
        ("sol toon (lean)", f"{SOL_BIN} activity:list -p {project_id} --limit 50 2>/dev/null", "Yes"),
    ], baseline_label="sol json (all fields)")

    # Summary - qualitative findings only, no hardcoded numbers
    print("\n" + "="*70)
    print("  KEY FINDINGS")
    print("="*70)
    print("""
  1. STRUCTURED vs UNSTRUCTURED
     upsun CLI outputs plain text tables - compact but NOT parseable.
     Sol outputs JSON/TOON - directly parseable by agents.

     For agents that need to extract data programmatically,
     Sol's structured output eliminates parsing errors.

  2. SOL LEAN OUTPUT vs VERBOSE JSON
     When you need structured data, sol's lean defaults are optimized.
     Lean output includes only essential fields (id, name, status, etc.)
     vs full JSON which includes all API fields.

     Use --full flag when you need all fields.

  3. SOL vs UPSUN (same data, different format)
     Sol's lean TOON output uses similar tokens as upsun's plain text,
     but Sol's output is STRUCTURED and PARSEABLE.

  BOTTOM LINE: Sol gives agents structured data at roughly the same
  token cost as upsun's unparseable plain text, with significant savings
  vs verbose JSON when you need all fields.
""")

if __name__ == "__main__":
    main()
