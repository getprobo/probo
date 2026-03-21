package trivy

default ignore = false

# @probo/vendors - internal package, CC-BY-SA-4.0 license is acceptable
ignore {
    input.PkgName == "@probo/vendors"
    input.Name == "CC-BY-SA-4.0"
}
