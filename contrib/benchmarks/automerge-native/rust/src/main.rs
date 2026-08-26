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

use std::collections::HashMap;
use std::fs;
use std::hint::black_box;
use std::time::Instant;

use automerge::transaction::{CommitOptions, Transactable};
use automerge::{ActorId, AutoCommit, ObjType, ReadDoc, ScalarValue, Value, ROOT};
use sha2::{Digest, Sha256};

struct BenchmarkWorkload {
    run: Box<dyn FnMut() -> Result<(), String>>,
    validate: Box<dyn FnMut() -> Result<String, String>>,
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

    println!(
        "{}",
        serde_json::json!({
            "workload": workload,
            "size": size,
            "iterations": iterations,
            "totalNs": total_ns,
            "nsPerOp": ns_per_op,
            "checksum": checksum,
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
        }),
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
        }),
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
            })
        }
        "save" => {
            let data = fixture_bytes(size, fixture)?;
            let mut document = AutoCommit::load(&data).map_err(|error| error.to_string())?;
            Ok(BenchmarkWorkload {
                run: Box::new(move || {
                    black_box(document.save());
                    Ok(())
                }),
                validate: Box::new(move || {
                    let mut document =
                        AutoCommit::load(&data).map_err(|error| error.to_string())?;
                    text_checksum(&mut document)
                }),
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

fn fixture_data(size: usize) -> Result<Vec<u8>, String> {
    let mut document = new_document();
    let text = document
        .put_object(&ROOT, "body", ObjType::Text)
        .map_err(|error| error.to_string())?;
    document
        .splice_text(&text, 0, 0, &benchmark_text(size))
        .map_err(|error| error.to_string())?;
    commit(&mut document)?;

    Ok(document.save())
}

fn fixture_bytes(size: usize, file: Option<&str>) -> Result<Vec<u8>, String> {
    match file {
        Some(file) => fs::read(file).map_err(|error| error.to_string()),
        None => fixture_data(size),
    }
}

fn new_document() -> AutoCommit {
    AutoCommit::new().with_actor(ActorId::from((0_u8..16_u8).collect::<Vec<_>>()))
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

fn checksum(value: &[u8]) -> String {
    hex::encode(Sha256::digest(value))
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
