# Code Collector: C++ to Go Conversions for QAgent Testing

This document collects C++ problem solutions, converts them to Go, and adds them to QAgent's testdata for diverse test coverage.

---

## 1. Graph Problem: Connected Components (Forest)

**Source:** CodeForces 755C - PolandBall and Forest (DFS-based graph problem)

**C++ Code:**
```cpp
#include <bits/stdc++.h>
using namespace std;

vector<vector<int>> adj_list;
vector<bool> visited;

void TreeExploration(int x) {
    visited[x] = true;
    for (auto y : adj_list[x]) {
        if (visited[y] == false) {
            TreeExploration(y);
        }
    }
}

int main() {
    int n; cin >> n;
    adj_list.resize(n+1);
    visited.resize(n+1);
    fill(visited.begin(), visited.end(), false);
    
    for (int i = 1; i < n+1; i++) {
        int x; cin >> x;
        adj_list[x].push_back(i);
        adj_list[i].push_back(x);
    }
    
    int cnt = 0;
    for (int i = 1; i < n+1; i++) {
        if (visited[i] == false) {
            TreeExploration(i);
            cnt++;
        }
    }
    cout << cnt;
}
```

**Challenges for Go Tests:**
- Multiple branches in loop (if/else on visited check)
- Recursive function edge cases
- Slice operations (append, iteration)
- Boundary conditions (1-based indexing)

**Go Equivalent Location:** `testdata/graph/poland_ball.go`

#####

## 2. Tree Problem: Maximum Depth

**Source:** CodeForces 115A - Party (Tree traversal, finding max depth)

**C++ Code:**
```cpp
#include <bits/stdc++.h>
using namespace std;

vector<vector<int>> adj_list;

int dfs(int x) {
    int max_depth = 0;
    for (auto v : adj_list[x]) {
        int child_depth = dfs(v);
        if (child_depth > max_depth) {
            max_depth = child_depth;
        }
    }
    return max_depth + 1;
}

int main(){
    int n; cin >> n;
    adj_list.resize(n+1);
    vector<int> managers;
    
    for (int i = 1; i < n+1; i++) {
        int x; cin >> x;
        if (x == -1) {
            managers.push_back(i);
        } else {
            adj_list[x].push_back(i);
        }
    }
    
    int max_depth = 0;
    for (auto x : managers) {
        int dist = dfs(x);
        if (dist > max_depth) {
            max_depth = dist;
        }
    }
    cout << max_depth << endl;
}
```

**Challenges for Go Tests:**
- Error handling for -1 sentinel values
- Recursion depth tracking
- Multiple root nodes (managers)
- Comparison and max tracking logic

**Go Equivalent Location:** `testdata/tree/party.go`

#####

## 3. Dynamic Programming: Modular Arithmetic

**Source:** Fibonacci variant with modular arithmetic

**C++ Code:**
```cpp
#include <iostream>
#include <map>
using namespace std;
using ll = long long;
const int M = 1000000007;

map<ll, ll> dp;

ll f(ll n) {
    if (n <= 2) return n >= 1;
    if (dp.count(n)) return dp[n];
    
    auto a = f(n / 2);
    auto b = f(n / 2 + 1);
    
    if (n % 2 == 0) {
        dp[n] = (2 * b - a) * a;
    } else {
        dp[n] = a * a + b * b;
    }
    dp[n] = (dp[n] % M + M) % M;
    return dp[n];
}

int main() {
    ll n;
    cin >> n;
    cout << f(n) << "\n";
}
```

**Challenges for Go Tests:**
- Memoization with maps
- Modular arithmetic (%, +)
- Conditional logic (if/else for even/odd)
- Recursive calls with cache lookup
- Edge cases (n <= 2)

**Go Equivalent Location:** `testdata/dp/fibonacci_mod.go`

#####

## 4. Buggy Code: Intentional Errors for Healing

**Source:** Teaching example with deliberate bugs

**C++ Code:**
```cpp
#include <iostream>
using namespace std;

int Add_subtract(int first, int second) {
    cout << "Add_subtract() received " << first << second << "\n";
    return (first+second);
    return (first-second);  // UNREACHABLE CODE
}

int main() {
    using std::cout;
    using std::cin;
    cout << "I'm in main()!\n";
    
    int a, b, c;
    cout << "enter 2 numbers";
    cin >> a;
    cin >> b;
    cout << "\nCalling Add_subtract()\n";
    
    c = Add_subtract(a,b);
    cout << "\nBack in main().\n";
    cout << "c was set to " << c;
    return 0;
}
```

**Bugs (for Go version):**
- Unreachable code (multiple returns)
- Logic errors (both operations, only first executes)
- I/O without proper formatting

**Go Equivalent Location:** `testdata/buggy/buggy_math.go`

#####

## Summary: Test File Locations

| File | Purpose | Error Type Expected |
|------|---------|-------------------|
| `testdata/graph/poland_ball.go` | DFS, graph traversal | runtime_error (nil pointer), low_coverage |
| `testdata/tree/party.go` | Tree depth, recursion | low_coverage (missing edge cases) |
| `testdata/dp/fibonacci_mod.go` | DP memoization | compilation (map usage), low_coverage |
| `testdata/buggy/buggy_math.go` | Intentional bugs | syntax_error, unreachable code |
| `testdata/parser/string_parser.go` | String parsing | missing_import, runtime_error |
| `testdata/calculator/complex_calc.go` | Multi-branch logic | low_coverage |
| `testdata/collections/ops.go` | Array/map operations | runtime_error (bounds, nil) |
| `testdata/validator/input.go` | Input validation | compilation, low_coverage |
| `testdata/statemachine/state.go` | State transitions | low_coverage (all states) |

---

**Total Test Files:** 9 functions across diverse problem domains
**Expected to trigger:** All 6 error categories (missing_import, syntax_error, compilation, runtime_error, low_coverage, unknown)
