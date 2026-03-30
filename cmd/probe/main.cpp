#include <iostream>
#include <string>
#include <vector>
#include <thread>
#include <mutex>
#include <atomic>
#include <cstring>
#include <netdb.h>
#include <arpa/inet.h>
#include <netinet/ip.h>
#include <netinet/tcp.h>
#include <sys/socket.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>
#include <signal.h>

using namespace std;

class LuxProbe {
private:
    string target;
    vector<int> open_ports;
    mutex port_mutex;
    atomic<int> active_threads{0};
    atomic<bool> running{true};
    
    bool is_host_up(const string& host) {
        struct addrinfo hints, *res;
        memset(&hints, 0, sizeof(hints));
        hints.ai_family = AF_UNSPEC;
        hints.ai_socktype = SOCK_STREAM;
        if (getaddrinfo(host.c_str(), NULL, &hints, &res) != 0) return false;
        freeaddrinfo(res);
        return true;
    }
    
    vector<string> resolve_dns(const string& host) {
        vector<string> ips;
        struct addrinfo hints, *res, *p;
        memset(&hints, 0, sizeof(hints));
        hints.ai_family = AF_UNSPEC;
        hints.ai_socktype = SOCK_STREAM;
        if (getaddrinfo(host.c_str(), NULL, &hints, &res) == 0) {
            for (p = res; p != NULL; p = p->ai_next) {
                char ip[INET6_ADDRSTRLEN];
                if (p->ai_family == AF_INET) {
                    struct sockaddr_in *ipv4 = (struct sockaddr_in*)p->ai_addr;
                    inet_ntop(AF_INET, &(ipv4->sin_addr), ip, INET_ADDRSTRLEN);
                    ips.push_back(string(ip));
                } else if (p->ai_family == AF_INET6) {
                    struct sockaddr_in6 *ipv6 = (struct sockaddr_in6*)p->ai_addr;
                    inet_ntop(AF_INET6, &(ipv6->sin6_addr), ip, INET6_ADDRSTRLEN);
                    ips.push_back(string(ip));
                }
            }
            freeaddrinfo(res);
        }
        return ips;
    }
    
    bool check_port(const string& host, int port, int timeout_sec = 2) {
        int sock = socket(AF_INET, SOCK_STREAM, 0);
        if (sock < 0) return false;
        
        struct timeval timeout;
        timeout.tv_sec = timeout_sec;
        timeout.tv_usec = 0;
        setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, &timeout, sizeof(timeout));
        setsockopt(sock, SOL_SOCKET, SO_SNDTIMEO, &timeout, sizeof(timeout));
        
        struct sockaddr_in addr;
        addr.sin_family = AF_INET;
        addr.sin_port = htons(port);
        inet_pton(AF_INET, host.c_str(), &addr.sin_addr);
        
        int flags = fcntl(sock, F_GETFL, 0);
        fcntl(sock, F_SETFL, flags | O_NONBLOCK);
        connect(sock, (struct sockaddr*)&addr, sizeof(addr));
        
        fd_set fdset;
        FD_ZERO(&fdset);
        FD_SET(sock, &fdset);
        
        struct timeval tv;
        tv.tv_sec = timeout_sec;
        tv.tv_usec = 0;
        
        bool open = false;
        if (select(sock + 1, NULL, &fdset, NULL, &tv) == 1) {
            int so_error;
            socklen_t len = sizeof(so_error);
            getsockopt(sock, SOL_SOCKET, SO_ERROR, &so_error, &len);
            if (so_error == 0) open = true;
        }
        close(sock);
        return open;
    }
    
    string grab_banner(const string& host, int port, int timeout_sec = 3) {
        int sock = socket(AF_INET, SOCK_STREAM, 0);
        if (sock < 0) return "";
        
        struct timeval timeout;
        timeout.tv_sec = timeout_sec;
        timeout.tv_usec = 0;
        setsockopt(sock, SOL_SOCKET, SO_RCVTIMEO, &timeout, sizeof(timeout));
        setsockopt(sock, SOL_SOCKET, SO_SNDTIMEO, &timeout, sizeof(timeout));
        
        struct sockaddr_in addr;
        addr.sin_family = AF_INET;
        addr.sin_port = htons(port);
        inet_pton(AF_INET, host.c_str(), &addr.sin_addr);
        
        if (connect(sock, (struct sockaddr*)&addr, sizeof(addr)) != 0) {
            close(sock);
            return "";
        }
        
        string probe;
        if (port == 80 || port == 8080) probe = "GET / HTTP/1.0\r\nHost: " + host + "\r\n\r\n";
        else if (port == 443 || port == 8443) probe = "HEAD / HTTP/1.0\r\n\r\n";
        else if (port == 22) probe = "SSH-2.0-Client\r\n";
        else if (port == 21) probe = "HELP\r\n";
        else probe = "\r\n";
        
        send(sock, probe.c_str(), probe.length(), 0);
        char buffer[4096];
        int bytes = recv(sock, buffer, sizeof(buffer) - 1, 0);
        close(sock);
        
        if (bytes > 0) {
            buffer[bytes] = '\0';
            string result(buffer);
            result.erase(remove(result.begin(), result.end(), '\n'), result.end());
            result.erase(remove(result.begin(), result.end(), '\r'), result.end());
            if (result.length() > 200) result = result.substr(0, 200);
            return result;
        }
        return "";
    }
    
    void scan_ports_range(const string& host, int start, int end) {
        active_threads++;
        for (int port = start; port <= end && running; port++) {
            if (check_port(host, port, 1)) {
                string banner = grab_banner(host, port, 2);
                lock_guard<mutex> lock(port_mutex);
                open_ports.push_back(port);
                cout << "PORT_OPEN " << port << "|" << banner << endl;
            }
        }
        active_threads--;
    }
    
public:
    LuxProbe(const string& target_host) : target(target_host) {}
    void stop() { running = false; }
    
    void scan(int start_port = 1, int end_port = 1024, int thread_count = 50) {
        if (!is_host_up(target)) {
            cout << "HOST_DOWN" << endl;
            return;
        }
        vector<string> ips = resolve_dns(target);
        cout << "DNS_RESULTS" << endl;
        for (const auto& ip : ips) cout << ip << endl;
        string main_ip = ips.empty() ? target : ips[0];
        cout << "SCAN_START " << main_ip << " " << start_port << " " << end_port << endl;
        int ports_per_thread = (end_port - start_port + 1) / thread_count;
        vector<thread> threads;
        for (int i = 0; i < thread_count; i++) {
            int thread_start = start_port + (i * ports_per_thread);
            int thread_end = (i == thread_count - 1) ? end_port : start_port + ((i + 1) * ports_per_thread) - 1;
            threads.emplace_back(&LuxProbe::scan_ports_range, this, main_ip, thread_start, thread_end);
        }
        for (auto& t : threads) t.join();
        cout << "SCAN_COMPLETE " << open_ports.size() << endl;
    }
};

LuxProbe* g_probe = nullptr;
void signal_handler(int sig) { if (g_probe) g_probe->stop(); exit(0); }

int main(int argc, char** argv) {
    if (argc < 2) { cout << "Usage: lux_probe <target> [start_port] [end_port]" << endl; return 1; }
    signal(SIGINT, signal_handler); signal(SIGTERM, signal_handler);
    string target = argv[1];
    int start = (argc > 2) ? atoi(argv[2]) : 1;
    int end = (argc > 3) ? atoi(argv[3]) : 1024;
    LuxProbe probe(target);
    g_probe = &probe;
    probe.scan(start, end);
    return 0;
}
