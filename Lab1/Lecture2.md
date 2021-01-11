# Lecture 2

#### If an inner function that is defined inside an outer function uses a local variable created by the outer function, and the outer function returns before the inner function ends, what will happen?
* If this happens, then the compiler for this chunk of code in Golang will automatically create this local variable inside heap, instead of stack. Both outer and inner function will refer to the variable on heap, and garbage collector in Golang will free this memory after the reference bit to this variable becomes 0.
