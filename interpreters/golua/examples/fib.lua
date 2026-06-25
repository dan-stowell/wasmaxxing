-- fib.lua: exercises the golua interpreter running on wasm.
-- Prints the first N Fibonacci numbers (N from arg[1], default 10).
local n = tonumber(arg and arg[1]) or 10
local function fibs(count)
  local out, a, b = {}, 0, 1
  for _ = 1, count do
    out[#out + 1] = tostring(a)
    a, b = b, a + b
  end
  return out
end
print("first " .. n .. " Fibonacci numbers via Lua on wasm:")
print(table.concat(fibs(n), " "))
