// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

use std::env;
use std::fs;
use std::path::Path;

use benchmark_battery::{
    Automerge, big_paste_doc, deep_history_doc, list_splice_100, maps_in_maps_doc,
    poorly_simulated_typing_doc, text_splice_100,
};

fn main() {
    let output = env::args()
        .nth(1)
        .expect("usage: probo-automerge-official-fixtures OUTPUT_DIR");
    let output = Path::new(&output);

    fs::create_dir_all(output).expect("cannot create fixture directory");

    write(output, "big-paste-100000.automerge", big_paste_doc(100_000));
    write(
        output,
        "text-splice-100-100000.automerge",
        text_splice_100(100_000),
    );
    write(
        output,
        "list-splice-100-100000.automerge",
        list_splice_100(100_000),
    );
    write(
        output,
        "typing-10000.automerge",
        poorly_simulated_typing_doc(10_000),
    );
    write(
        output,
        "deep-history-1000.automerge",
        deep_history_doc(1_000),
    );
    write(
        output,
        "maps-in-maps-1000.automerge",
        maps_in_maps_doc(1_000),
    );
}

fn write(output: &Path, name: &str, document: Automerge) {
    fs::write(output.join(name), document.save()).expect("cannot write fixture");
}
