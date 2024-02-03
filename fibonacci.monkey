let fibonacci = fn(x) {
    if (x < 2) {
        return x;
    }
    fibonacci(x-1) + fibonacci(x-2);
}
