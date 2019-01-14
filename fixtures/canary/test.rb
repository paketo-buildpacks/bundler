stdout, $stdout = $stdout, $stderr
begin
  def data()
    puts Bundler::Dsl.evaluate("Gemfile", 'Gemfile.lock', {})::VERSION
    puts "______________________________"
    puts Bundler::VERSION
    return Bundler::VERSION
  end
  out = data()
  stdout.puts({error:nil, data:out}.to_json)
rescue => e
  stdout.puts({error:e.to_s, data:nil}.to_json)
end
