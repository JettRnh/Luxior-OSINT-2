use std::fs;
use std::path::Path;
use std::collections::HashSet;
use rayon::prelude::*;
use serde::{Serialize, Deserialize};
use regex::Regex;
use ignore::Walk;

#[derive(Debug, Serialize, Deserialize)]
pub struct ParsedData {
    pub source_file: String,
    pub emails: Vec<String>,
    pub urls: Vec<String>,
    pub ip_addresses: Vec<String>,
    pub phone_numbers: Vec<String>,
    pub crypto_addresses: Vec<String>,
    pub social_media: Vec<String>,
}

pub struct Parser {
    email_regex: Regex,
    url_regex: Regex,
    ip_regex: Regex,
    phone_regex: Regex,
    btc_regex: Regex,
    eth_regex: Regex,
    twitter_regex: Regex,
    github_regex: Regex,
}

impl Parser {
    pub fn new() -> Self {
        Parser {
            email_regex: Regex::new(r"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}").unwrap(),
            url_regex: Regex::new(r"https?://[a-zA-Z0-9.-]+(/[a-zA-Z0-9._~:/?#[\]@!$&'()*+,;=]*)?").unwrap(),
            ip_regex: Regex::new(r"\b(?:\d{1,3}\.){3}\d{1,3}\b").unwrap(),
            phone_regex: Regex::new(r"(\+?[0-9]{1,3}[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}").unwrap(),
            btc_regex: Regex::new(r"\b[13][a-km-zA-HJ-NP-Z1-9]{25,34}\b").unwrap(),
            eth_regex: Regex::new(r"\b0x[a-fA-F0-9]{40}\b").unwrap(),
            twitter_regex: Regex::new(r"(?:twitter\.com/|@)([a-zA-Z0-9_]{1,15})").unwrap(),
            github_regex: Regex::new(r"(?:github\.com/)([a-zA-Z0-9_-]+)").unwrap(),
        }
    }
    
    pub fn parse_file(&self, path: &Path, content: &str) -> ParsedData {
        ParsedData {
            source_file: path.to_string_lossy().to_string(),
            emails: self.extract_unique(&self.email_regex, content),
            urls: self.extract_unique(&self.url_regex, content),
            ip_addresses: self.extract_unique(&self.ip_regex, content),
            phone_numbers: self.extract_unique(&self.phone_regex, content),
            crypto_addresses: self.collect_crypto(content),
            social_media: self.collect_social(content),
        }
    }
    
    fn extract_unique(&self, regex: &Regex, content: &str) -> Vec<String> {
        let mut set = HashSet::new();
        for cap in regex.captures_iter(content) {
            if let Some(m) = cap.get(0) { set.insert(m.as_str().to_string()); }
        }
        set.into_iter().collect()
    }
    
    fn collect_crypto(&self, content: &str) -> Vec<String> {
        let mut crypto = Vec::new();
        for cap in self.btc_regex.captures_iter(content) {
            if let Some(m) = cap.get(0) { crypto.push(format!("BTC:{}", m.as_str())); }
        }
        for cap in self.eth_regex.captures_iter(content) {
            if let Some(m) = cap.get(0) { crypto.push(format!("ETH:{}", m.as_str())); }
        }
        crypto.sort(); crypto.dedup(); crypto
    }
    
    fn collect_social(&self, content: &str) -> Vec<String> {
        let mut social = Vec::new();
        for cap in self.twitter_regex.captures_iter(content) {
            if let Some(m) = cap.get(1) { social.push(format!("Twitter:@{}", m.as_str())); }
        }
        for cap in self.github_regex.captures_iter(content) {
            if let Some(m) = cap.get(1) { social.push(format!("GitHub:{}", m.as_str())); }
        }
        social.sort(); social.dedup(); social
    }
    
    pub fn parse_directory(&self, dir_path: &str) -> Vec<ParsedData> {
        let files: Vec<_> = Walk::new(dir_path)
            .filter_map(|e| e.ok())
            .filter(|e| e.path().is_file())
            .map(|e| e.path().to_path_buf())
            .collect();
        files.par_iter().filter_map(|p| {
            fs::read_to_string(p).ok().map(|c| self.parse_file(p, &c))
        }).collect()
    }
}

fn main() {
    let args: Vec<_> = std::env::args().collect();
    if args.len() < 2 { eprintln!("Usage: lux_parser <dir>"); std::process::exit(1); }
    let parser = Parser::new();
    let results = parser.parse_directory(&args[1]);
    let json = serde_json::to_string_pretty(&results).unwrap();
    fs::write("lux_parser_output.json", json).unwrap();
    println!("Parsed {} files", results.len());
      }
