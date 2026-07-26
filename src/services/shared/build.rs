use std::path::Path;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    println!("cargo:rerun-if-changed=proto/");

    let proto_root = Path::new("proto");

    // Collect all .proto files recursively
    let proto_files = collect_proto_files(proto_root);

    if proto_files.is_empty() {
        println!("cargo:warning=No .proto files found in proto/ directory");
        return Ok(());
    }

    // Configure tonic_build — compile all .proto files into OUT_DIR
    tonic_build::configure()
        .build_server(true)
        .build_client(true)
        .compile_protos(&proto_files, &[proto_root])?;

    println!("cargo:info=Compiled {} .proto files", proto_files.len());
    Ok(())
}

/// Recursively collect all .proto files under a directory.
fn collect_proto_files(dir: &Path) -> Vec<std::path::PathBuf> {
    let mut files = Vec::new();
    if !dir.exists() {
        return files;
    }
    for entry in std::fs::read_dir(dir).expect("failed to read proto directory") {
        let entry = entry.expect("failed to read directory entry");
        let path = entry.path();
        if path.is_dir() {
            files.extend(collect_proto_files(&path));
        } else if path.extension().is_some_and(|e| e == "proto") {
            files.push(path);
        }
    }
    files
}
