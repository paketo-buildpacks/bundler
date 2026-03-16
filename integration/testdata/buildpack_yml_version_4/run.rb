# minimal ruby only server for integration testing
require "socket"
require "open3"

port = 8080
server = TCPServer.new port

while session = server.accept
  begin
    request = session.gets
    puts request

    session.print "HTTP/1.1 200\r\n" # 1
    session.print "Content-Type: text/html\r\n" # 2
    session.print "\r\n" # 3

    [
      "which bundler",
      "bundle version",
      "which ruby",
      "ruby --version"
    ].each do |command|
      output, _ = Open3.capture2e(command)
      session.print "$ #{command}\n#{output}\n"
    end

    session.close

  rescue Errno::EPIPE
    puts "Connection broke!"
  end
end
