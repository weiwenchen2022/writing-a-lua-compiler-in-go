local mymod = require "mymod"
mymod.foo()
mymod.bar()

print "\npackage.config:"
print(package.config)

print("\npackage.path: " .. package.path)

print "\npackage.loaded:"
for k, v in pairs(package.loaded) do print(k, v) end

print "\npackage.preload:"
for k, v in pairs(package.preload) do print(k, v) end

print "\npackage.searchers:"
for i, v in ipairs(package.searchers) do print(i, v) end