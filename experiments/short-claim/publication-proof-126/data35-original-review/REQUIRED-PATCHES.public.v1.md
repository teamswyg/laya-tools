# Required source corrections

Apply only in a Root-owned fresh stage or immutable corrected successor; preserve the reviewed author seal and frozen inputs.

## 1. Count project Reader error returns

In source/cmd/main.go.txt load, change the initial Reader error branch:

```go
ex, e := shortclaimdata.LoadDevelopmentRow(bytes.NewReader(raw))
if e != nil {
    return data35.Loaded{ReaderReturned: true}, errors.New("reader_failed")
}
```

Keep ValidateAttempted=false on this branch. A failed call has one explicit dispatch and one observed return; it has zero matching rows and zero extra bridge validation calls. Add or extend a synthetic bridge-boundary control using a fake injected Reader/pure return-state helper; do not invoke the project Reader merely to construct a fake control. Existing library controls already accept returned=true plus a subsequent error, but they do not exercise this CLI branch. Root independently confirms the patched bridge source and appropriate controls.

## 2. Sync empty reserved files before input reads

After both O_EXCL files open successfully in Reserve and before its syncDir(out)/return:

```go
if r.data.Sync() != nil {
    return r, code("data_reservation_sync")
}
if r.receipt.Sync() != nil {
    return r, code("receipt_reservation_sync")
}
if e = syncDir(out); e != nil {
    return r, e
}
```

CLI's existing defer reservation.Close() must still run for a nonnil partial Reservation returned with an error. Keep all created files/directories; no unlink, clobber, reopen or retry. A fake sync-failure control should require error propagation and zero subsequent Load/adoption/Reader calls; the normal control requires both empty file syncs before Reserve returns. Retain parent-directory syncing after mkdir and data/receipt syncing on Finish.

These are accounting and durability corrections. They change no frozen input, Want, caption, order, label, role, selected-U metadata or qualification. Root alone formats, compiles, tests and executes the corrected source. This note is source advice; neither patch nor any control was executed by the reviewer.
