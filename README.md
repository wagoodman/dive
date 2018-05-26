# docker-image-explorer
```
docker build -t die-test:latest .
go run main.go
```

# TODO:

- [x] Extract docker layers from api
- [x] Represent layers as generic tree
- [x] Stack ordere tree list together as fake unionfs tree
- [ ] Diff trees
- [ ] Add ui for browsing layers
- [ ] Add ui for diffing stack to layer
