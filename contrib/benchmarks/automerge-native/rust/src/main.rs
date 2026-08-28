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

use std::cell::RefCell;
use std::collections::HashMap;
use std::fs;
use std::hint::black_box;
use std::rc::Rc;
use std::time::Instant;

use automerge::sync::{self, SyncDoc};
use automerge::transaction::{CommitOptions, Transactable};
use automerge::{ActorId, AutoCommit, ObjType, ReadDoc, ScalarValue, Value, ROOT};
use sha2::{Digest, Sha256};

struct BenchmarkWorkload {
    run: Box<dyn FnMut() -> Result<(), String>>,
    validate: Box<dyn FnMut() -> Result<String, String>>,
    output: Box<dyn FnMut() -> Option<Vec<u8>>>,
}

fn main() {
    let arguments = arguments().unwrap_or_else(|error| fail(&error));
    let workload = arguments
        .get("workload")
        .cloned()
        .unwrap_or_else(|| fail("workload is required"));
    let size = parse_usize(&arguments, "size", 0);
    let iterations = parse_usize(&arguments, "iterations", 1);
    let warmups = parse_usize(&arguments, "warmups", 3);
    let fixture = arguments.get("fixture").map(String::as_str);
    if iterations == 0 {
        fail("iterations must be positive");
    }

    let mut runner = workload_runner(&workload, size, fixture).unwrap_or_else(|error| fail(&error));
    for _ in 0..warmups {
        (runner.run)().unwrap_or_else(|error| fail(&error));
    }

    let started_at = Instant::now();
    for _ in 0..iterations {
        (runner.run)().unwrap_or_else(|error| fail(&error));
    }
    let total_ns = started_at.elapsed().as_nanos();
    let ns_per_op = total_ns / iterations as u128;
    let checksum = (runner.validate)().unwrap_or_else(|error| fail(&error));
    let output = (runner.output)();
    let output_bytes = output.as_ref().map_or(0, Vec::len);
    let output_hash = output.as_deref().map(checksum_bytes);

    println!(
        "{}",
        serde_json::json!({
            "workload": workload,
            "size": size,
            "iterations": iterations,
            "totalNs": total_ns,
            "nsPerOp": ns_per_op,
            "checksum": checksum,
            "outputBytes": output_bytes,
            "outputHash": output_hash,
        })
    );
}

fn workload_runner(
    workload: &str,
    size: usize,
    fixture: Option<&str>,
) -> Result<BenchmarkWorkload, String> {
    match workload {
        "create" => Ok(BenchmarkWorkload {
            run: Box::new(|| {
                let document = new_document();
                black_box(&document);
                Ok(())
            }),
            validate: Box::new(|| Ok(checksum(b"empty"))),
            output: Box::new(|| None),
        }),
        "map" => Ok(BenchmarkWorkload {
            run: Box::new(move || {
                let document = map_document(size)?;
                black_box(&document);
                Ok(())
            }),
            validate: Box::new(move || {
                let mut document = map_document(size)?;
                map_checksum(&mut document, size)
            }),
            output: Box::new(|| None),
        }),
        "map-update" => {
            let document = map_document(size)?;
            let (_, values) = document
                .get(&ROOT, "values")
                .map_err(|error| error.to_string())?
                .ok_or_else(|| "values map does not exist".to_owned())?;
            let document = Rc::new(RefCell::new(document));
            let run_document = Rc::clone(&document);
            Ok(BenchmarkWorkload {
                run: Box::new(move || {
                    let mut document = run_document.borrow_mut();
                    for index in 0..size {
                        document
                            .put(&values, index.to_string(), (size - index) as i64)
                            .map_err(|error| error.to_string())?;
                    }
                    commit(&mut document)
                }),
                validate: Box::new(move || map_checksum(&mut document.borrow_mut(), size)),
                output: Box::new(|| None),
            })
        }
        "text" => Ok(BenchmarkWorkload {
            run: Box::new(move || {
                let document = typed_document(size)?;
                black_box(&document);
                Ok(())
            }),
            validate: Box::new(move || {
                let mut document = typed_document(size)?;
                text_checksum(&mut document)
            }),
            output: Box::new(|| None),
        }),
        "text-edit" => {
            let document = fixture_document(size)?;
            let (_, text) = document
                .get(&ROOT, "body")
                .map_err(|error| error.to_string())?
                .ok_or_else(|| "body text does not exist".to_owned())?;
            let document = Rc::new(RefCell::new(document));
            let run_document = Rc::clone(&document);
            Ok(BenchmarkWorkload {
                run: Box::new(move || {
                    let mut document = run_document.borrow_mut();
                    let edits = size.min(100);
                    for index in 0..edits {
                        let position = index * size / edits;
                        document
                            .splice_text(&text, position, 1, "z")
                            .map_err(|error| error.to_string())?;
                    }
                    commit(&mut document)
                }),
                validate: Box::new(move || text_checksum(&mut document.borrow_mut())),
                output: Box::new(|| None),
            })
        }
        "load" => {
            let data = fixture_bytes(size, fixture)?;
            let validation_data = data.clone();
            Ok(BenchmarkWorkload {
                run: Box::new(move || {
                    let document = AutoCommit::load(&data).map_err(|error| error.to_string())?;
                    black_box(&document);
                    Ok(())
                }),
                validate: Box::new(move || {
                    let mut document =
                        AutoCommit::load(&validation_data).map_err(|error| error.to_string())?;
                    text_checksum(&mut document)
                }),
                output: Box::new(|| None),
            })
        }
        "save" => {
            let data = fixture_bytes(size, fixture)?;
            let mut document = AutoCommit::load(&data).map_err(|error| error.to_string())?;
            let latest_save = Rc::new(RefCell::new(None));
            let run_save = Rc::clone(&latest_save);
            let validation_save = Rc::clone(&latest_save);
            let output_save = Rc::clone(&latest_save);
            Ok(BenchmarkWorkload {
                run: Box::new(move || {
                    let data = black_box(document.save());
                    *run_save.borrow_mut() = Some(data);
                    Ok(())
                }),
                validate: Box::new(move || {
                    let data = validation_save.borrow();
                    let data = data
                        .as_deref()
                        .ok_or_else(|| "save workload did not produce data".to_owned())?;
                    let mut document = AutoCommit::load(data).map_err(|error| error.to_string())?;
                    text_checksum(&mut document)
                }),
                output: Box::new(move || output_save.borrow().clone()),
            })
        }
        "save-after-loaded-change" | "save-after-change" => {
            let data = fixture_bytes(size, fixture)?;
            let mut document = AutoCommit::load(&data).map_err(|error| error.to_string())?;
            let (_, text) = document
                .get(&ROOT, "body")
                .map_err(|error| error.to_string())?
                .ok_or_else(|| "body text does not exist".to_owned())?;
            let latest_save = Rc::new(RefCell::new(None));
            let run_save = Rc::clone(&latest_save);
            let validation_save = Rc::clone(&latest_save);
            let output_save = Rc::clone(&latest_save);
            let mut position = size;
            Ok(BenchmarkWorkload {
                run: Box::new(move || {
                    document
                        .splice_text(&text, position, 0, "x")
                        .map_err(|error| error.to_string())?;
                    position += 1;
                    commit(&mut document)?;
                    let data = black_box(document.save());
                    *run_save.borrow_mut() = Some(data);
                    Ok(())
                }),
                validate: Box::new(move || {
                    let data = validation_save.borrow();
                    let data = data
                        .as_deref()
                        .ok_or_else(|| "save workload did not produce data".to_owned())?;
                    let mut document = AutoCommit::load(data).map_err(|error| error.to_string())?;
                    text_checksum(&mut document)
                }),
                output: Box::new(move || output_save.borrow().clone()),
            })
        }
        "merge-loaded" | "merge-reloaded" | "concurrent-tail-reconcile" => {
            let tail_edits = if workload == "concurrent-tail-reconcile" {
                size.min(100)
            } else {
                1
            };
            let (left_data, right_data) = merge_fixture_data(size, tail_edits)?;
            if workload == "merge-reloaded" {
                let validation_left = left_data.clone();
                let validation_right = right_data.clone();
                return Ok(BenchmarkWorkload {
                    run: Box::new(move || {
                        let document = merge_documents(&left_data, &right_data)?;
                        black_box(&document);
                        Ok(())
                    }),
                    validate: Box::new(move || {
                        let mut document = merge_documents(&validation_left, &validation_right)?;
                        text_checksum(&mut document)
                    }),
                    output: Box::new(|| None),
                });
            }

            let left = AutoCommit::load(&left_data).map_err(|error| error.to_string())?;
            let right = AutoCommit::load(&right_data).map_err(|error| error.to_string())?;
            let left = Rc::new(RefCell::new(left));
            let right = Rc::new(RefCell::new(right));
            let run_left = Rc::clone(&left);
            let run_right = Rc::clone(&right);
            Ok(BenchmarkWorkload {
                run: Box::new(move || {
                    run_left
                        .borrow_mut()
                        .merge(&mut run_right.borrow_mut())
                        .map_err(|error| error.to_string())?;
                    Ok(())
                }),
                validate: Box::new(move || text_checksum(&mut left.borrow_mut())),
                output: Box::new(|| None),
            })
        }
        "sync-initial" => {
            let data = fixture_bytes(size, fixture)?;
            let latest = Rc::new(RefCell::new(None));
            let run_latest = Rc::clone(&latest);
            let latest_wire = Rc::new(RefCell::new(None));
            let run_wire = Rc::clone(&latest_wire);
            let output_wire = Rc::clone(&latest_wire);
            Ok(BenchmarkWorkload {
                run: Box::new(move || {
                    let mut left = AutoCommit::load(&data)
                        .map_err(|error| error.to_string())?
                        .with_actor(benchmark_actor());
                    let mut right = new_document().with_actor(benchmark_peer_actor());
                    let mut left_state = sync::State::new();
                    let mut right_state = sync::State::new();
                    let wire =
                        synchronize_wire(&mut left, &mut right, &mut left_state, &mut right_state)?;
                    *run_wire.borrow_mut() = Some(wire);
                    *run_latest.borrow_mut() = Some(right);
                    Ok(())
                }),
                validate: Box::new(move || {
                    let mut latest = latest.borrow_mut();
                    let document = latest.as_mut().ok_or_else(|| {
                        "initial sync workload did not produce a document".to_owned()
                    })?;
                    text_checksum(document)
                }),
                output: Box::new(move || output_wire.borrow().clone()),
            })
        }
        "sync-diverged" => {
            let data = fixture_bytes(size, fixture)?;
            let mut left = AutoCommit::load(&data)
                .map_err(|error| error.to_string())?
                .with_actor(benchmark_actor());
            let mut right = AutoCommit::load(&data)
                .map_err(|error| error.to_string())?
                .with_actor(benchmark_peer_actor());
            let mut left_state = sync::State::new();
            let mut right_state = sync::State::new();
            synchronize(&mut left, &mut right, &mut left_state, &mut right_state)?;

            let (_, left_text) = left
                .get(&ROOT, "body")
                .map_err(|error| error.to_string())?
                .ok_or_else(|| "left body text does not exist".to_owned())?;
            left.splice_text(&left_text, size, 0, "L")
                .map_err(|error| error.to_string())?;
            commit(&mut left)?;

            let (_, right_text) = right
                .get(&ROOT, "body")
                .map_err(|error| error.to_string())?
                .ok_or_else(|| "right body text does not exist".to_owned())?;
            right
                .splice_text(&right_text, size, 0, "R")
                .map_err(|error| error.to_string())?;
            commit(&mut right)?;

            let left = Rc::new(RefCell::new(left));
            let right = Rc::new(RefCell::new(right));
            let run_left = Rc::clone(&left);
            let run_right = Rc::clone(&right);
            let latest_wire = Rc::new(RefCell::new(None));
            let run_wire = Rc::clone(&latest_wire);
            let output_wire = Rc::clone(&latest_wire);
            Ok(BenchmarkWorkload {
                run: Box::new(move || {
                    let wire = synchronize_wire(
                        &mut run_left.borrow_mut(),
                        &mut run_right.borrow_mut(),
                        &mut left_state,
                        &mut right_state,
                    )?;
                    *run_wire.borrow_mut() = Some(wire);
                    Ok(())
                }),
                validate: Box::new(move || {
                    let left_checksum = text_checksum(&mut left.borrow_mut())?;
                    let right_checksum = text_checksum(&mut right.borrow_mut())?;
                    if left_checksum != right_checksum {
                        return Err("synchronized documents differ".to_owned());
                    }
                    Ok(left_checksum)
                }),
                output: Box::new(move || output_wire.borrow().clone()),
            })
        }
        other => Err(format!("unknown workload {other:?}")),
    }
}

fn map_document(size: usize) -> Result<AutoCommit, String> {
    let mut document = new_document();
    let values = document
        .put_object(&ROOT, "values", ObjType::Map)
        .map_err(|error| error.to_string())?;
    for index in 0..size {
        document
            .put(&values, index.to_string(), index as i64)
            .map_err(|error| error.to_string())?;
    }
    commit(&mut document)?;

    Ok(document)
}

fn typed_document(size: usize) -> Result<AutoCommit, String> {
    let mut document = new_document();
    let text = document
        .put_object(&ROOT, "body", ObjType::Text)
        .map_err(|error| error.to_string())?;
    for index in 0..size {
        document
            .splice_text(&text, index, 0, "x")
            .map_err(|error| error.to_string())?;
    }
    commit(&mut document)?;

    Ok(document)
}

fn fixture_document(size: usize) -> Result<AutoCommit, String> {
    let mut document = new_document();
    let text = document
        .put_object(&ROOT, "body", ObjType::Text)
        .map_err(|error| error.to_string())?;
    document
        .splice_text(&text, 0, 0, &benchmark_text(size))
        .map_err(|error| error.to_string())?;
    commit(&mut document)?;

    Ok(document)
}

fn fixture_data(size: usize) -> Result<Vec<u8>, String> {
    Ok(fixture_document(size)?.save())
}

fn fixture_bytes(size: usize, file: Option<&str>) -> Result<Vec<u8>, String> {
    match file {
        Some(file) => fs::read(file).map_err(|error| error.to_string()),
        None => fixture_data(size),
    }
}

fn merge_fixture_data(size: usize, tail_edits: usize) -> Result<(Vec<u8>, Vec<u8>), String> {
    let mut base = new_document();
    let text = base
        .put_object(&ROOT, "body", ObjType::Text)
        .map_err(|error| error.to_string())?;
    base.splice_text(&text, 0, 0, &benchmark_text(size))
        .map_err(|error| error.to_string())?;
    commit(&mut base)?;
    let base_data = base.save();

    let mut left = AutoCommit::load(&base_data)
        .map_err(|error| error.to_string())?
        .with_actor(benchmark_actor());
    let (_, left_text) = left
        .get(&ROOT, "body")
        .map_err(|error| error.to_string())?
        .ok_or_else(|| "left body text does not exist".to_owned())?;
    for index in 0..tail_edits {
        left.splice_text(&left_text, size + index, 0, "L")
            .map_err(|error| error.to_string())?;
    }
    commit(&mut left)?;

    let mut right = AutoCommit::load(&base_data)
        .map_err(|error| error.to_string())?
        .with_actor(benchmark_peer_actor());
    let (_, right_text) = right
        .get(&ROOT, "body")
        .map_err(|error| error.to_string())?
        .ok_or_else(|| "right body text does not exist".to_owned())?;
    for index in 0..tail_edits {
        right
            .splice_text(&right_text, size + index, 0, "R")
            .map_err(|error| error.to_string())?;
    }
    commit(&mut right)?;

    Ok((left.save(), right.save()))
}

fn merge_documents(left_data: &[u8], right_data: &[u8]) -> Result<AutoCommit, String> {
    let mut left = AutoCommit::load(left_data).map_err(|error| error.to_string())?;
    let mut right = AutoCommit::load(right_data).map_err(|error| error.to_string())?;
    left.merge(&mut right).map_err(|error| error.to_string())?;

    Ok(left)
}

fn new_document() -> AutoCommit {
    AutoCommit::new().with_actor(benchmark_actor())
}

fn benchmark_actor() -> ActorId {
    ActorId::from((0_u8..16_u8).collect::<Vec<_>>())
}

fn benchmark_peer_actor() -> ActorId {
    ActorId::from((0_u8..16_u8).rev().collect::<Vec<_>>())
}

fn commit(document: &mut AutoCommit) -> Result<(), String> {
    document
        .commit_with(
            CommitOptions::default()
                .with_message("benchmark")
                .with_time(0),
        )
        .ok_or_else(|| "change contains no operations".to_owned())?;

    Ok(())
}

fn benchmark_text(size: usize) -> String {
    (0..size)
        .map(|index| char::from(b'a' + (index % 26) as u8))
        .collect()
}

fn map_checksum(document: &mut AutoCommit, size: usize) -> Result<String, String> {
    let (_, values) = document
        .get(&ROOT, "values")
        .map_err(|error| error.to_string())?
        .ok_or_else(|| "values map does not exist".to_owned())?;
    let mut normalized = Vec::with_capacity(size * 16);
    for index in 0..size {
        let (value, _) = document
            .get(&values, index.to_string())
            .map_err(|error| error.to_string())?
            .ok_or_else(|| format!("map value {index} does not exist"))?;
        let Value::Scalar(value) = value else {
            return Err(format!("map value {index} is not a scalar"));
        };
        let ScalarValue::Int(value) = value.as_ref() else {
            return Err(format!("map value {index} is not an integer"));
        };
        normalized.extend_from_slice(value.to_string().as_bytes());
        normalized.push(b'\n');
    }

    Ok(checksum(&normalized))
}

fn text_checksum(document: &mut AutoCommit) -> Result<String, String> {
    let (_, text) = document
        .get(&ROOT, "body")
        .map_err(|error| error.to_string())?
        .ok_or_else(|| "body text does not exist".to_owned())?;
    let value = document.text(&text).map_err(|error| error.to_string())?;

    Ok(checksum(value.as_bytes()))
}

fn synchronize(
    left: &mut AutoCommit,
    right: &mut AutoCommit,
    left_state: &mut sync::State,
    right_state: &mut sync::State,
) -> Result<(), String> {
    synchronize_wire(left, right, left_state, right_state).map(|_| ())
}

fn synchronize_wire(
    left: &mut AutoCommit,
    right: &mut AutoCommit,
    left_state: &mut sync::State,
    right_state: &mut sync::State,
) -> Result<Vec<u8>, String> {
    let mut wire = Vec::new();
    for _ in 0..100 {
        let mut progressed = false;
        if let Some(message) = left.sync().generate_sync_message(left_state) {
            wire.extend_from_slice(&message.clone().encode());
            right
                .sync()
                .receive_sync_message(right_state, message)
                .map_err(|error| error.to_string())?;
            progressed = true;
        }
        if let Some(message) = right.sync().generate_sync_message(right_state) {
            wire.extend_from_slice(&message.clone().encode());
            left.sync()
                .receive_sync_message(left_state, message)
                .map_err(|error| error.to_string())?;
            progressed = true;
        }
        if !progressed {
            return Ok(wire);
        }
    }

    Err("sync did not quiesce".to_owned())
}

fn checksum(value: &[u8]) -> String {
    hex::encode(Sha256::digest(value))
}

fn checksum_bytes(value: &[u8]) -> String {
    checksum(value)
}

fn arguments() -> Result<HashMap<String, String>, String> {
    let mut values = HashMap::new();
    let mut arguments = std::env::args().skip(1);
    while let Some(argument) = arguments.next() {
        let name = argument
            .strip_prefix("--")
            .ok_or_else(|| format!("invalid argument {argument:?}"))?;
        let value = arguments
            .next()
            .ok_or_else(|| format!("missing value for {argument:?}"))?;
        values.insert(name.to_owned(), value);
    }
    Ok(values)
}

fn parse_usize(arguments: &HashMap<String, String>, name: &str, fallback: usize) -> usize {
    arguments
        .get(name)
        .map(|value| {
            value
                .parse()
                .unwrap_or_else(|_| fail(&format!("{name} must be an integer")))
        })
        .unwrap_or(fallback)
}

fn fail(message: &str) -> ! {
    eprintln!("{message}");
    std::process::exit(1);
}
