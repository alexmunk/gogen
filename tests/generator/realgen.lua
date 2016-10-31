function getline(t, i)
  count = 0
  for k,v in pairs(t) do
    if count >= i then
      return v
    else 
      count = count + 1
    end
  end
end

function countlines(t)
  count = 0
  for k,v in pairs(t) do
      count = count + 1
  end
  return count
end

function table.val_to_str ( v )
  if "string" == type( v ) then
    v = string.gsub( v, "\n", "\\n" )
    if string.match( string.gsub(v,"[^'\"]",""), '^"+$' ) then
      return "'" .. v .. "'"
    end
    return '"' .. string.gsub(v,'"', '\\"' ) .. '"'
  else
    return "table" == type( v ) and table.tostring( v ) or
      tostring( v )
  end
end

function table.key_to_str ( k )
  if "string" == type( k ) and string.match( k, "^[_%a][_%a%d]*$" ) then
    return k
  else
    return "[" .. table.val_to_str( k ) .. "]"
  end
end

function table.tostring( tbl )
  local result, done = {}, {}
  for k, v in ipairs( tbl ) do
    table.insert( result, table.val_to_str( v ) )
    done[ k ] = true
  end
  for k, v in pairs( tbl ) do
    if not done[ k ] then
      table.insert( result,
        table.key_to_str( k ) .. "=" .. table.val_to_str( v ) )
    end
  end
  return "{" .. table.concat( result, "," ) .. "}"
end

debug("cur: " .. state["cur"])
l = getline(lines, tonumber(state["cur"]))
l = replaceTokens(l)
ret = { }
table.insert(ret, l)
debug("ret: " .. table.tostring(ret))
state["cur"] = tonumber(state["cur"]) + 1
debug("cur: " .. state["cur"])
count = countlines(lines)
if state["cur"] >= count then
  state["cur"] = 0
end
return ret