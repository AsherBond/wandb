// use sentry;
use std::collections::HashMap;
// use std::io;

use colored::*;

use std::thread;
use std::time::Duration;
use std::{cmp::min, fmt::Write};

use indicatif::{ProgressBar, ProgressState, ProgressStyle};

use wandbinder::session;

fn hyperlink(text: &str, url: &str) -> ColoredString {
    format!("\x1B]8;;{}\x07{}\x1B]8;;\x07", url, text).white()
}

fn get_prefix() -> ColoredString {
    String::from("│ ").truecolor(250, 193, 60)
}

fn get_checkmark() -> ColoredString {
    String::from("✓").truecolor(122, 166, 56)
}

// fn get_crossmark() -> ColoredString {
//     String::from("✗").truecolor(227, 50, 79)
// }

fn get_header() -> ColoredString {
    String::from("wandb").white().bold()
}

fn print_header(name: &str, url: &str) {
    let link = hyperlink(name, url);

    let prefix = get_prefix();
    let checkmark = get_checkmark();
    let header = get_header();

    println!("{}{}", prefix, header);

    // spinner
    let pb = ProgressBar::new_spinner();
    pb.enable_steady_tick(Duration::from_millis(120));
    pb.set_prefix(prefix.to_string());
    pb.set_style(
        ProgressStyle::with_template("{prefix}{spinner:.magenta} {msg}")
            .unwrap()
            // For more spinners check out the cli-spinners project:
            // https://github.com/sindresorhus/cli-spinners/blob/master/spinners.json
            .tick_strings(&["⠋", "⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓"]),
    );
    pb.set_message("Creating run...");
    thread::sleep(Duration::from_secs(3));

    // pb.finish_with_message("Done");
    pb.finish_and_clear();

    println!("{}{} Run created - {}", prefix, checkmark, link);
    // println!("{}{} {}", prefix, crossmark, dimmed);
    // println!();
}

fn print_footer(name: &str, url: &str) {
    let link = hyperlink(name, url);

    let mut downloaded = 0;
    let total_size = 231231231;

    let prefix = get_prefix();
    let checkmark = get_checkmark();
    let header = get_header();

    println!("{}{}", prefix, header);

    let pb = ProgressBar::new(total_size);
    pb.set_prefix(prefix.to_string());
    pb.set_style(
        ProgressStyle::with_template(
            "{prefix}Syncing run {wide_bar:.magenta/white.dim} {bytes}/{total_bytes} ({eta})",
        )
        .unwrap()
        .with_key("eta", |state: &ProgressState, w: &mut dyn Write| {
            write!(w, "{:.1}s", state.eta().as_secs_f64()).unwrap()
        })
        // .progress_chars("⠿⠇"),
        .progress_chars("⣿⡇"),
    );

    while downloaded < total_size {
        let new = min(downloaded + 223211, total_size);
        downloaded = new;
        pb.set_position(new);
        thread::sleep(Duration::from_millis(12));
    }

    pb.finish_and_clear();

    println!("{}{} Run synced - {}", prefix, checkmark, link);
}

fn main() {
    // let _guard = sentry::init(
    //     "https://9e9d0694aa7ccd41aeb5bc34aadd716a@o151352.ingest.sentry.io/4506068829470720",
    // );
    // sentry::capture_error(&io::Error::new(io::ErrorKind::Other, "LOL HAI I AM ERROR"));

    let settings = session::Settings::new(None, Some(1.0), Some(1));

    let session = session::Session::new(settings);

    let mut run = session.init_run(None);

    let name = "glorious-capybara-23";
    let url = "https://wandb.ai/dimaduev/uncategorized/runs/KEHHBT";

    print_header(name, url);

    let mut data: HashMap<String, f64> = HashMap::new();
    data.insert("loss".to_string(), 13.37);

    run.log(data);
    println!("Logging to run {}...", run.id);
    thread::sleep(Duration::from_secs(2));

    run.finish();

    print_footer(name, url);
}