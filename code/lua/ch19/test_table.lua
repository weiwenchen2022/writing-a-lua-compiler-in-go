
a = {1, 2, 3, 4, 5,}
table.move(a, 1, #a, 2)
a[1] = 0
print(table.concat(a, ", "))

table.move(a, 2, #a, 1)
a[#a] = nil
print(table.concat(a, ", "))

b = table.move(a, 1, #a, 1, {})
print(table.concat(b, ", "))

table.move(a, 1, #a, #b + 1, b)
print(table.concat(b, ", "))