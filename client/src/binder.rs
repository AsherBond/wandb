// pub mod wandb_internal {
//     include!(concat!(env!("OUT_DIR"), "/wandb_internal.rs"));
// }

// use wandb_internal;
mod wandb_internal;

fn main() {
    println!("Hello, world!");
    println!(
        "{:?}",
        wandb_internal::Settings {
            base_url: Some("https://google.com".to_string()),
            ..Default::default()
        }
    );
    println!(
        "{:?}",
        wandb_internal::RequestInfo {
            ..Default::default()
        }
    );
    println!(
        "{:?}",
        wandb_internal::ServerShutdownRequest {
            ..Default::default()
        }
    )
}