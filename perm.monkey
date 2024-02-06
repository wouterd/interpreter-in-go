let combine = fn(a, b) {
    if (len(b) == 0) {
        return a
    }
    return combine(push(a,first(b)), rest(b))
}

let contains = fn(arr, el) {
    if (len(arr) == 0) {
        return false
    }
    if (first(arr) == el) {
        return true
    }
    return contains(rest(arr), el)
}

let for = fn(curr, stop, f) {
    let acc = []
    if (curr < stop) {
        let acc = f(acc, curr)
        let acc = combine(acc, for(curr+1, stop, f))
    }
    return acc
}

let nums = for(1, 8, fn(acc, x) { push(acc, x) })

let backtrack = fn(solution) {
    if (len(nums) == len(solution)) {
        return [solution]
    } else {
        return for (0, len(nums), fn(acc, x) { 
            if (!contains(solution, nums[x])) {
                return combine(acc, backtrack(push(solution, nums[x])))
            }
            return acc
        })
    } 
}

puts(backtrack([]))
