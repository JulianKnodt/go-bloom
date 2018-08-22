# Bloom

Bloom is a simple threadsafe bloom filter written in go.

It is fairly simple and only exports `Insert`, `PossiblyContains`, and `Len`

It only functions on fixed size data structures, so no slices, maps, or int/uint where the size of the type
is unspecified
