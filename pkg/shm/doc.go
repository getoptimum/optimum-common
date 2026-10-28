// Package shm is the client for rlnc-server's shared-memory lanes. Each
// operation claims a lane, writes a length-prefixed OpInput frame into the
// lane's input region, wakes the server, and reads the OpOutput frame back from
// the output region before releasing the lane.
//
// # Hand-written envelope encoding
//
// OpInput and OpOutput frames are encoded and parsed with protowire, not with
// proto.Marshal and proto.Unmarshal. The request is marshaled straight into the
// mapped input region and the response is parsed in place from the output
// region. *That removes one staging copy on each side. Decode requests carry a
// chunk's whole symbol set, so those copies dominated SHM allocations.*
//
// The cost is that this package hard-codes the envelope schema owned by
// buf.build/getoptimum/rlnc (generated into pkg/rlncpb/shm). writeInputFrame
// emits exactly the bytes proto.Marshal would, including proto3's omission of
// zero-valued fields. parseOpOutput accepts exactly what proto.Unmarshal would.
//
// Regenerating pkg/rlncpb is not enough when OpInput or OpOutput changes!!!
// When regenerating pkg/rlncpb, we must ensure:
//
//   - Update the field-number constants, writeInputFrame, and parseOpOutput.
//   - A new OpInput field is never sent, and a new OpOutput field is skipped as
//     unknown, until this package handles it.
//   - Run the frame tests. TestOpFrameFieldNumbersMatchSchema fails on a
//     renumbered or renamed field. The round-trip tests fail on encoding drift.
package shm
