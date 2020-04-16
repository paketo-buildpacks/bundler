# minimal ruby only server for integration testing
require "socket"
require "open3"

port = 8080
server = TCPServer.new port

while session = server.accept
  request = session.gets
  puts request

  session.print "HTTP/1.1 200\r\n" # 1
  session.print "Content-Type: text/html\r\n" # 2
  session.print "\r\n" # 3

  session.print "Hello World!\n"

  stdout, stderr, status = Open3.capture3("bundle version")
  session.print "status: #{status}\n"
  session.print "stdout:\n#{stdout}"

  if !stderr.empty?
    session.print "stderr:\n#{stderr}"
  end

  session.close
end
