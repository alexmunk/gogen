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

if state["stage"] == 0 then
    -- Pick a user
    
end