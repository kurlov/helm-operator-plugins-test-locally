### Run

1. Install CRD 
```kubectl apply -f crd.yaml```

2. Install CR
```kubectl apply -f nginx-cr.yaml```

3. Run reconciler
```go run main.go nginx```