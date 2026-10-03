package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

type session struct {
	cfg         config
	out         string
	controlUsed int64
	journal     *os.File
	seq         int
	previous    string
	result      outcome
	// Owned fake controls only. Production run never sets this hook.
	fault func(stage string, targetIndex, fileIndex int) error
}

func (s *session) charge(n int64) error {
	// Keep room for one bounded stdout result as well as retained controls.
	if n < 0 || n > controlCap-2048-s.controlUsed {
		return errCap
	}
	s.controlUsed += n
	return nil
}
func (s *session) writeControl(path string, v any) (pin, error) {
	b, e := canonical(v)
	if e != nil || len(b) > 16384 {
		return pin{}, errCap
	}
	if e := s.charge(int64(len(b))); e != nil {
		return pin{}, e
	}
	if e := noSymlinkAbsolute(filepath.Dir(path)); e != nil {
		return pin{}, e
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return pin{}, e
	}
	n, we := f.Write(b)
	if we == nil && n != len(b) {
		we = io.ErrShortWrite
	}
	if we == nil {
		we = f.Sync()
	}
	ce := f.Close()
	if we != nil {
		return pin{}, we
	}
	if ce != nil {
		return pin{}, ce
	}
	if e := syncDir(filepath.Dir(path)); e != nil {
		return pin{}, e
	}
	return pin{path, int64(len(b)), digest(b)}, nil
}

type event struct {
	Seq         int    `json:"s"`
	Kind        int    `json:"k"`
	Target      int    `json:"t"`
	FileIndex   int    `json:"i"`
	PreviousSHA string `json:"p"`
	CommitSHA   string `json:"c,omitempty"`
}

func (s *session) appendEvent(kind, ti, fi int, commitSHA string) error {
	b, e := json.Marshal(event{s.seq + 1, kind, ti, fi, s.previous, commitSHA})
	if e != nil {
		return e
	}
	b = append(b, '\n')
	if e := s.charge(int64(len(b))); e != nil {
		return e
	}
	n, e := s.journal.Write(b)
	if e == nil && n != len(b) {
		e = io.ErrShortWrite
	}
	if e == nil {
		e = s.journal.Sync()
	}
	if e == nil {
		e = syncDir(s.out)
	}
	if e != nil {
		return e
	}
	s.seq++
	s.previous = digest(b)
	return nil
}

func removeSelectedTree(root string, t target, original bool, s *session, ti int) error {
	if e := checkTree(root, t, !original, original); e != nil {
		return e
	}
	// Capture inode facts after full verification. Re-check these exact objects
	// immediately before removal; a same-content replacement is not authorized.
	rootInfo, e := os.Lstat(root)
	if e != nil {
		return e
	}
	type physical struct {
		info  os.FileInfo
		entry entry
	}
	captured := make([]physical, len(t.Files))
	for i, x := range t.Files {
		p := filepath.Join(root, filepath.FromSlash(x.Relative))
		fi, e := os.Lstat(p)
		if e != nil {
			return e
		}
		captured[i] = physical{fi, x}
	}
	directoryInfos := map[string]os.FileInfo{}
	for _, d := range t.Directories {
		fi, e := os.Lstat(filepath.Join(root, filepath.FromSlash(d.Relative)))
		if e != nil {
			return e
		}
		directoryInfos[d.Relative] = fi
	}
	for i, cap := range captured {
		if s.fault != nil {
			if e := s.fault("before_remove", ti, i); e != nil {
				return e
			}
		}
		x := cap.entry
		p := filepath.Join(root, filepath.FromSlash(x.Relative))
		got, before, e := hashRegular(p, restoreCap)
		if e != nil || got.Bytes != x.Bytes || got.SHA256 != x.SHA256 || !os.SameFile(cap.info, before) {
			return errChanged
		}
		cm, e := infoMeta(cap.info)
		if e != nil || !sameMeta(before, cm, true) {
			return errChanged
		}
		if original {
			if !sameMeta(before, x.meta, true) {
				return errChanged
			}
			if e := s.appendEvent(1, ti, i, ""); e != nil {
				return e
			}
		}
		if e := noSymlinkAbsolute(filepath.Dir(p)); e != nil {
			return e
		}
		last, e := os.Lstat(p)
		if e != nil || !sameMeta(last, cm, true) {
			return errChanged
		}
		if e := os.Remove(p); e != nil {
			return e
		}
		if original {
			s.result.UnacknowledgedRemovePossible = true
		}
		if e := syncDir(filepath.Dir(p)); e != nil {
			return e
		}
		if original {
			if e := s.appendEvent(2, ti, i, ""); e != nil {
				return e
			}
			s.result.RemovedACKFiles++
			s.result.RemovedACKLogicalBytes += x.Bytes
			s.result.UnacknowledgedRemovePossible = false
		}
	}
	for _, d := range orderedDirectories(t, true) {
		p := filepath.Join(root, filepath.FromSlash(d.Relative))
		if e := noSymlinkAbsolute(p); e != nil {
			return e
		}
		fi, e := os.Lstat(p)
		if e != nil || !fi.IsDir() || !os.SameFile(fi, directoryInfos[d.Relative]) {
			return errChanged
		}
		children, e := os.ReadDir(p)
		if e != nil || len(children) != 0 {
			return errArtifact
		}
		if e := os.Remove(p); e != nil {
			return e
		}
		if e := syncDir(filepath.Dir(p)); e != nil {
			return e
		}
	}
	current, e := os.Lstat(root)
	if e != nil || !os.SameFile(rootInfo, current) {
		return errChanged
	}
	remaining, e := os.ReadDir(root)
	if e != nil {
		return e
	}
	if original {
		if len(remaining) != 1 || remaining[0].Name() != markerName {
			return errArtifact
		}
		return syncDir(root)
	}
	if len(remaining) != 0 {
		return errArtifact
	}
	if e := os.Remove(root); e != nil {
		return e
	}
	return syncDir(filepath.Dir(root))
}

func executeTarget(s *session, sel selection, a admission, ti int) error {
	t := sel.Targets[ti]
	source := filepath.Join(s.cfg.Root, t.ID)
	restore := filepath.Join(s.cfg.Root, s.cfg.RestoreBases[ti])
	archive := filepath.Join(s.out, strconv.Itoa(ti)+".tar.gz")
	ap, e := archiveTarget(source, archive, t, a.ArchiveCaps[ti])
	if e != nil {
		return e
	}
	archiveBytes := ap.Bytes
	s.result.ArchiveBytes[ti] = &archiveBytes
	if e := restoreArchive(ap, restore, t); e != nil {
		return e
	}
	// All exact regular paths, empty files and modes were physically restored;
	// a decoder-only roundtrip is never substituted for this proof.
	cp, e := s.writeControl(filepath.Join(s.out, "COMMIT."+strconv.Itoa(ti)+".json"), commit{
		Schema: "riido-hf33-directory-recoverable-commit-v1", TargetID: t.ID,
		Selection: s.cfg.Selection, Admission: s.cfg.Admission, PriorAccount: s.cfg.PriorAccount,
		InventorySHA: inventorySHA(t), Archive: ap, RestoredRoot: restore,
		PayloadBytes: payloadSizes[ti], RegularFiles: len(t.Files),
		FullEOFCRCVerified: true, RestoredFilesBytesSHAAndModesMatch: true,
		FileAndDirectorySyncReturned: true, BeforeAnyOriginalUnlink: true,
	})
	if e != nil {
		return e
	}
	if e := s.appendEvent(0, ti, -1, cp.SHA256); e != nil {
		return e
	}
	// Root remains an accounting root. Only this exact fresh marker is added.
	marker, e := s.writeControl(filepath.Join(source, markerName), struct {
		Schema    string `json:"schema"`
		Commit    pin    `json:"commit"`
		Selection pin    `json:"selection"`
		Archive   pin    `json:"archive"`
	}{"riido-hf33-directory-recoverable-marker-v1", cp, s.cfg.Selection, ap})
	if e != nil {
		return e
	}
	if _, e := readPin(cp, 16384); e != nil {
		return e
	}
	if _, e := readPin(marker, 16384); e != nil {
		return e
	}
	if actual, _, e := hashRegular(ap.Path, maxArchiveCap); e != nil || actual != ap {
		return errChanged
	}
	if e := removeSelectedTree(source, t, true, s, ti); e != nil {
		return e
	}

	// Temp cleanup is permitted only after durable COMMIT + original cleanup.
	// A failure here preserves any remaining restore files, with complete=false.
	if e := removeSelectedTree(restore, t, false, s, ti); e != nil {
		return e
	}
	if _, e := os.Lstat(restore); !os.IsNotExist(e) {
		return errArtifact
	}
	if e := syncDir(s.cfg.Root); e != nil {
		return e
	}
	if _, e := readPin(cp, 16384); e != nil {
		return e
	}
	if _, e := readPin(marker, 16384); e != nil {
		return e
	}
	if actual, _, e := hashRegular(ap.Path, maxArchiveCap); e != nil || actual != ap {
		return errChanged
	}
	ack, e := s.writeControl(filepath.Join(s.out, "CLEANUP-ACK."+strconv.Itoa(ti)+".json"), struct {
		Schema                 string `json:"schema"`
		Commit                 pin    `json:"commit"`
		Marker                 pin    `json:"marker"`
		Archive                pin    `json:"archive"`
		OriginalRegularsAbsent bool   `json:"original_regulars_absent"`
		OriginalRootMarkerOnly bool   `json:"original_root_marker_only"`
		RestoredSiblingAbsent  bool   `json:"restored_sibling_absent"`
		ParentSyncReturned     bool   `json:"parent_sync_returned"`
	}{"riido-hf33-directory-recoverable-cleanup-ack-v1", cp, marker, ap, true, true, true, true})
	if e != nil {
		return e
	}
	if e := s.appendEvent(3, ti, -1, ack.SHA256); e != nil {
		return e
	}
	s.result.CompletedTargets++
	return nil
}

func run(configPath, configSHA string) (result outcome, runErr error) {
	result = outcome{Schema: "riido-hf33-directory-archive-result-v1"}
	if !validSHA(configSHA) {
		result.Failure = "config_pin"
		return result, errContract
	}
	p, _, e := hashRegular(configPath, 16384)
	if e != nil || p.SHA256 != configSHA {
		result.Failure = "config_pin"
		return result, errChanged
	}
	raw, e := readPin(p, 16384)
	if e != nil {
		result.Failure = "config_pin"
		return result, e
	}
	var c config
	if e := decodeCanonical(raw, &c); e != nil {
		result.Failure = "config_canonical"
		return result, e
	}
	sb, e := readPin(c.Selection, metadataCap)
	if e != nil {
		result.Failure = "selection_pin"
		return result, e
	}
	var sel selection
	if e := decodeCanonical(sb, &sel); e != nil {
		result.Failure = "selection_canonical"
		return result, e
	}
	if e := checkSelection(sel); e != nil {
		result.Failure = "selection_contract"
		return result, e
	}
	ab, e := readPin(c.Admission, 16384)
	if e != nil {
		result.Failure = "admission_pin"
		return result, e
	}
	var a admission
	if e := decodeCanonical(ab, &a); e != nil {
		result.Failure = "admission_canonical"
		return result, e
	}
	if e := checkAdmission(c, a); e != nil {
		result.Failure = "admission_or_peak"
		return result, e
	}
	if e := checkPaths(c, configPath, a); e != nil {
		result.Failure = "physical_paths"
		return result, e
	}
	for _, t := range sel.Targets {
		if e := checkTree(filepath.Join(c.Root, t.ID), t, false, false); e != nil {
			result.Failure = "source_inventory"
			return result, e
		}
	}
	out := filepath.Join(c.Root, c.OutputBase)
	if e := os.Mkdir(out, 0700); e != nil {
		result.Failure = "exclusive_output"
		return result, e
	}
	if e := syncDir(c.Root); e != nil {
		result.Failure = "output_parent_sync"
		return result, e
	}
	s := &session{cfg: c, out: out, result: result}
	defer func() {
		if s.journal != nil {
			s.journal.Close()
		}
		// Failure emits only a bounded control summary. It never cleans files,
		// removes originals automatically, retries, or asserts reclaimed bytes.
		if runErr != nil {
			s.result.Complete = false
			s.result.ReclaimedLogicalBytes = nil
			if s.result.Failure == "" {
				s.result.Failure = "archive_incomplete"
			}
			s.writeControl(filepath.Join(out, "FAILURE.json"), s.result)
		}
		result = s.result
	}()
	reservation, e := s.writeControl(filepath.Join(out, "RESERVATION.json"), struct {
		Schema            string  `json:"schema"`
		Config            pin     `json:"config"`
		Selection         pin     `json:"selection"`
		Admission         pin     `json:"admission"`
		Cap               int64   `json:"cap_bytes"`
		ArchiveCaps       []int64 `json:"archive_caps"`
		RestorePayloadCap int64   `json:"restore_payload_cap_bytes"`
		ControlCap        int64   `json:"new_control_cap_bytes"`
	}{"riido-hf33-directory-archive-reservation-v1", p, c.Selection, c.Admission, globalCap, a.ArchiveCaps, restoreCap, controlCap})
	if e != nil {
		return result, e
	}
	s.previous = reservation.SHA256
	s.journal, e = os.OpenFile(filepath.Join(out, "UNLINK-JOURNAL.jsonl"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return result, e
	}
	if e := s.journal.Sync(); e != nil {
		return result, e
	}
	if e := syncDir(out); e != nil {
		return result, e
	}
	for ti := 0; ti < 2; ti++ {
		// Frozen config/selection/admission remain immutable through both phases.
		if _, e := readPin(p, 16384); e != nil {
			return result, e
		}
		if _, e := readPin(c.Selection, metadataCap); e != nil {
			return result, e
		}
		if _, e := readPin(c.Admission, 16384); e != nil {
			return result, e
		}
		if e := executeTarget(s, sel, a, ti); e != nil {
			return result, e
		}
	}
	// Both CLEANUP ACKs exist before this final result is committed. A partial
	// output failure downgrades stdout instead of presenting a false complete.
	s.result.Complete = true
	if s.result.ArchiveBytes[0] == nil || s.result.ArchiveBytes[1] == nil {
		return result, errContract
	}
	net := s.result.RemovedACKLogicalBytes - *s.result.ArchiveBytes[0] - *s.result.ArchiveBytes[1] - controlCap
	s.result.ReclaimedLogicalBytes = &net // Conservative: entire new-control reserve.
	if _, e := s.writeControl(filepath.Join(out, "FINAL.json"), s.result); e != nil {
		return result, e
	}
	return s.result, nil
}
